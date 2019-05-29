/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package agentplugin

// Agent plugin registry

import (
	"context"
	"sync"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/pkg/plugin"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type AgentRegistry struct {
	*plugin.Registry
}

var defaultRegistry = NewAgentRegistry()

func NewAgentRegistry() *AgentRegistry {
	r := &AgentRegistry{plugin.NewRegistry()}

	return r
}

func (r *AgentRegistry) CollectHeartbeatData(ctx context.Context) (PluginHeartbeatDataList, map[plugin.PluginName]derrors.Error) {
	dataList := []PluginHeartbeatData{}
	errMap := make(map[plugin.PluginName]derrors.Error, len(r.Running()))
	timedout := false
	// Lock for collecting results
	var dataLock sync.Mutex

	// Wait for plugin Beats to execute in parallel
	var beatWaitGroup sync.WaitGroup

	// Collect data in parallel
	for name, p := range(r.Running()) {
		agentPlugin, ok := p.(AgentPlugin)
		if !ok {
			// Skip non-agent plugins in registry
			continue
		}
		beatWaitGroup.Add(1)
		// We need to copy the variables to something local to this
		// loop - the range will modify them and the running routines
		// will use the modified values otherwise
		go func(name plugin.PluginName, p AgentPlugin){
			defer beatWaitGroup.Done()
			data, err := p.Beat(ctx)

			dataLock.Lock()
			// Don't add more when we already timed out
			if !timedout {
				// Always set err so we know we've processed this plugin
				errMap[name] = err
				if data != nil && err == nil {
					dataList = append(dataList, data)
				}
			}
			dataLock.Unlock()
		}(name, agentPlugin)
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
		log.Warn().Msg("Timeout collecting plugin heartbeat data")
		dataLock.Lock()
		timedout = true
		dataLock.Unlock()

	}

	// Set errors for interrupted Beats
	for name := range(r.Running()) {
		err, found := errMap[name]
		if !found {
			if timedout {
				// Processing timed out
				log.Warn().Str("plugin", name.String()).Msg("Plugin timed out")
				errMap[name] = derrors.NewDeadlineExceededError("plugin beat timed out").WithParams(name.String())
			} else {
				// Should have been found!
				log.Warn().Str("plugin", name.String()).Msg("Plugin did not run while it should have")
				errMap[name] = derrors.NewAbortedError("plugin beat did not run").WithParams(name.String())
			}
		} else if err == nil {
			// Sanitize error map while we're at it
			delete(errMap, name)
		}
	}

	return dataList, errMap
}
func Register(p *plugin.PluginDescriptor) derrors.Error {
	return defaultRegistry.Register(p)
}

func ListPlugins() plugin.RegistryEntryMap {
	return defaultRegistry.ListPlugins()
}

func StartPlugin(name plugin.PluginName, conf *viper.Viper) derrors.Error {
	return defaultRegistry.StartPlugin(name, conf)
}

func StopPlugin(name plugin.PluginName) derrors.Error {
	return defaultRegistry.StopPlugin(name)
}

func StopAll() {
	defaultRegistry.StopAll()
}

func ExecuteCommand(ctx context.Context, name plugin.PluginName, cmd plugin.CommandName, params map[string]string) (string, derrors.Error) {
	return defaultRegistry.ExecuteCommand(ctx, name,cmd, params)
}

func CollectHeartbeatData(ctx context.Context) (PluginHeartbeatDataList, map[plugin.PluginName]derrors.Error) {
	return defaultRegistry.CollectHeartbeatData(ctx)
}
