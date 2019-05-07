/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package plugin

// Plugin registry

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Type for mapping plugin names to functions that create them
type AvailablePluginMap map[PluginName]*PluginDescriptor

// Type for mapping plugin names to running plugins
type RunningPluginMap map[PluginName]Plugin

type Registry struct {
	available AvailablePluginMap
	running RunningPluginMap
}

var defaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	r := &Registry{
		available: AvailablePluginMap{},
		running: RunningPluginMap{},
	}

	return r
}

func (r *Registry) Register(plugin *PluginDescriptor) derrors.Error {
	_, found := r.available[plugin.Name]
	if found {
		return derrors.NewInvalidArgumentError("plugin already registered").WithParams(plugin.Name)
	}

	r.available[plugin.Name] = plugin
	log.Debug().Str("name", plugin.Name.String()).Msg("plugin registered")
	return nil
}

func Register(plugin *PluginDescriptor) derrors.Error {
	return defaultRegistry.Register(plugin)
}

func (r *Registry) StartPlugin(name PluginName, conf *viper.Viper) (derrors.Error) {
	log.Debug().Str("name", name.String()).Interface("config", conf.AllSettings()).Msg("starting plugin")
	plugin, found := r.available[name]
	if !found {
		return derrors.NewInvalidArgumentError("plugin not available").WithParams(name)
	}

	_, found = r.running[name]
	if found {
		return derrors.NewInvalidArgumentError("plugin already running").WithParams(name)
	}

	instance, derr := plugin.NewFunc(conf)
	if derr != nil {
		return derr
	}

	derr = instance.StartPlugin()
	if derr != nil {
		return derr
	}

	r.running[name] = instance
	return nil
}

func StartPlugin(name PluginName, conf *viper.Viper) derrors.Error {
	return defaultRegistry.StartPlugin(name, conf)
}

func (r *Registry) StopPlugin(name PluginName) (derrors.Error) {
	log.Debug().Str("name", name.String()).Msg("stopping plugin")
	plugin, found := r.running[name]
	if !found {
		return derrors.NewInvalidArgumentError("plugin not running").WithParams(name)
	}

	delete(r.running, name)
	plugin.StopPlugin()

	return nil
}

func StopPlugin(name PluginName) derrors.Error {
	return defaultRegistry.StopPlugin(name)
}


func (r *Registry) ExecuteCommand(ctx context.Context, name PluginName, cmd CommandName, params map[string]string) (string, derrors.Error) {
	plugin, found := r.running[name]
	if !found {
		_, found := r.available[name]
		if !found {
			return "", derrors.NewInvalidArgumentError("plugin not available").WithParams(name)
		}
		return "", derrors.NewInvalidArgumentError("plugin not running").WithParams(name)
	}

	cmdFunc := plugin.GetCommandFunc(cmd)
	if cmdFunc == nil {
		return "", derrors.NewInvalidArgumentError("command not available").WithParams(name, cmd)
	}

	return cmdFunc(ctx, params)
}

func ExecuteCommand(ctx context.Context, name PluginName, cmd CommandName, params map[string]string) (string, derrors.Error) {
	return defaultRegistry.ExecuteCommand(ctx, name,cmd, params)
}