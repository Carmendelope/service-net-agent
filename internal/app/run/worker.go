/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Agent operations worker

import (
	"context"
	"fmt"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// We have one worker. Currently, we also have one thread dispatching work,
// so all operations are serialized. If we introduce multiple dispatcher
// threads, we need to ensure that operations are still in the right order;
// at least per plugin. We can either maintain a single worker with proper
// locking to ensure this, or dynamically create a worker per plugin and
// make sure the worker always executes in a serial manner. We also have to
// take care that a single plugin (with serialized actions) shouldn't block
// all (or actually, more than one) dispatcher thread.
type Worker struct {
	config *config.Config
}

func NewWorker(config *config.Config) *Worker {
	w := &Worker{
		config: config,
	}

	return w
}

func (w *Worker) Execute(ctx context.Context, name plugin.PluginName, cmd plugin.CommandName, params map[string]string) (string, derrors.Error) {
	var result string
	var derr derrors.Error = nil

	switch cmd {
	case plugin.StartCommand:
		config := createPluginConfig(params)
		derr = plugin.StartPlugin(name, config)
		if derr == nil {
			w.writePluginConfig(name, config)
		}
	case plugin.StopCommand:
		derr = plugin.StopPlugin(name)
	default:
		result, derr = plugin.ExecuteCommand(ctx, name, cmd, params)
	}

	return result, derr
}

func (w *Worker) writePluginConfig(name plugin.PluginName, config *viper.Viper) {
	confName := name.String()

	log.Debug().Str("name", confName).Msg("writing plugin configuration")
	for k, v := range(config.AllSettings()) {
		w.config.Set(fmt.Sprintf("%s.%s", confName, k), v)
	}
	// We always set something to make to config file entry exist
	w.config.Set(fmt.Sprintf("%s.enabled", confName), true)

	// This writes the plugin config, which also makes the plugin persistent on reboot
	derr := w.config.Write()
	if derr != nil {
		log.Warn().Err(derr).Str("trace", derr.DebugReport()).Str("name", name.String()).Msg("failed persisting plugin configuration")
	}
}

func createPluginConfig(params map[string]string) *viper.Viper {
	conf := viper.New()
	for k, v := range(params) {
		conf.Set(k, v)
	}

	return conf
}
