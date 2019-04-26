/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package plugin

// Plugin registry

import (
	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Type for mapping plugin names to functions that create them
type AvailablePluginMap map[PluginName]NewPluginFunc

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

func (r *Registry) Register(name PluginName, newFunc NewPluginFunc) derrors.Error {
	_, found := r.available[name]
	if found {
		return derrors.NewInvalidArgumentError("plugin already registered").WithParams(name)
	}

	r.available[name] = newFunc
	log.Debug().Str("name", name.String()).Msg("plugin registered")
	return nil
}

func Register(name PluginName, newFunc NewPluginFunc) derrors.Error {
	return defaultRegistry.Register(name, newFunc)
}

func (r *Registry) StartPlugin(name PluginName, conf *viper.Viper) (derrors.Error) {
	newFunc, found := r.available[name]
	if !found {
		return derrors.NewInvalidArgumentError("plugin not available").WithParams(name)
	}

	_, found = r.running[name]
	if found {
		return derrors.NewInvalidArgumentError("plugin already running").WithParams(name)
	}

	instance, derr := newFunc(conf)
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

// TBD StopPlugin

func (r *Registry) ExecuteCommand(name PluginName, cmd CommandName, params map[string]string) (string, derrors.Error) {
	plugin, found := r.running[name]
	if !found {
		_, found := r.available[name]
		if !found {
			return "", derrors.NewInvalidArgumentError("plugin not available").WithParams(name)
		}
		return "", derrors.NewInvalidArgumentError("plugin not running").WithParams(name)
	}

	cmdFunc := plugin.GetCommand(cmd)
	if cmdFunc == nil {
		return "", derrors.NewInvalidArgumentError("command not available").WithParams(name, cmd)
	}

	return cmdFunc(params)
}

func ExecuteCommand(name PluginName, cmd CommandName, params map[string]string) (string, derrors.Error) {
	return defaultRegistry.ExecuteCommand(name, cmd, params)
}
