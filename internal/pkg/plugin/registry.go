/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package plugin

// Plugin registry

import (
	"context"
	"sync"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Type for mapping plugin names to functions that create them
type AvailablePluginMap map[PluginName]*PluginDescriptor

// Type for mapping plugin names to running plugins
type RunningPluginMap map[PluginName]Plugin

// Dynamic type to list current status of plugins
type RegistryEntry struct {
	*PluginDescriptor
	Running bool
}

type RegistryEntryMap map[PluginName]RegistryEntry

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
	return nil
}

func Register(plugin *PluginDescriptor) derrors.Error {
	return defaultRegistry.Register(plugin)
}

func (r *Registry) ListPlugins() RegistryEntryMap {
	entries := make(RegistryEntryMap, len(r.available))

	for name, descriptor := range(r.available) {
		_, running := r.running[name]
		entries[name] = RegistryEntry{
			PluginDescriptor: descriptor,
			Running: running,
		}
	}

	return entries
}

func ListPlugins() RegistryEntryMap {
	return defaultRegistry.ListPlugins()
}

func (r *Registry) StartPlugin(name PluginName, conf *viper.Viper) (derrors.Error) {
	// Easier to handle nil configuration here
	if conf == nil {
		conf = viper.New()
	}

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

func (r *Registry) StopAll() {
	for name, _ := range(r.running) {
		r.StopPlugin(name)
	}
}

func StopAll() {
	defaultRegistry.StopAll()
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

func (r *Registry) CollectHeartbeatData(ctx context.Context) (PluginHeartbeatDataList, map[PluginName]derrors.Error) {
	dataList := []PluginHeartbeatData{}
	errMap := make(map[PluginName]derrors.Error, len(r.running))
	timedout := false
	// Lock for collecting results
	var dataLock sync.Mutex

	// Wait for plugin Beats to execute in parallel
	var beatWaitGroup sync.WaitGroup

	// Collect data in parallel
	for name, plugin := range(r.running) {
		beatWaitGroup.Add(1)
		go func(){
			defer beatWaitGroup.Done()
			data, err := plugin.Beat(ctx)

			dataLock.Lock()
			// Don't add more when we already timed out
			if !timedout {
				// Always set err so we know we've processed this plugin
				errMap[name] = err
				if err == nil {
					dataList = append(dataList, data)
				}
			}
			dataLock.Unlock()
		}()
	}

	// Interruptable wait
	doneChan := make(chan struct{})
	go func(){
		beatWaitGroup.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		// All good - everybody is done and we have our results
		break
	case <-ctx.Done():
		// Timeout - check which functions did time out and set errors
		// accordingly

		// Make sure the running goroutines don't add anything else
		// while we're collecting results
		dataLock.Lock()
		timedout = true
		dataLock.Unlock()

		// Set errors for interrupted Beats
		for name := range(r.running) {
			err, found := errMap[name]
			if !found {
				// Processing timed out
				errMap[name] = derrors.NewDeadlineExceededError("plugin beat timed out").WithParams(name.String())
			}

			// Sanitize error map while we're at it
			if err == nil {
				delete(errMap, name)
			}
		}

	}

	return dataList, errMap
}

func CollectHeartbeatData(ctx context.Context) (PluginHeartbeatDataList, map[PluginName]derrors.Error) {
	return defaultRegistry.CollectHeartbeatData(ctx)
}
