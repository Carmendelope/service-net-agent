/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Agent operations worker

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/spf13/viper"
)

type Worker struct {
}

func NewWorker() *Worker {
	w := &Worker{
	}

	return w
}

func (w *Worker) Execute(ctx context.Context, name plugin.PluginName, cmd plugin.CommandName, params map[string]string) (string, derrors.Error) {
	var result string
	var derr derrors.Error = nil

	switch cmd {
	case plugin.StartCommand:
		derr = plugin.StartPlugin(name, createPluginConfig(params))
	case plugin.StopCommand:
		derr = plugin.StopPlugin(name)
	default:
		result, derr = plugin.ExecuteCommand(ctx, name, cmd, params)
	}

	return result, derr
}

func createPluginConfig(params map[string]string) *viper.Viper {
	conf := viper.New()
	for k, v := range(params) {
		conf.Set(k, v)
	}

	return conf
}
