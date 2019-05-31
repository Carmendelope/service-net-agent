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
)

type AgentRegistry struct {
	*plugin.Registry
}

// This contains the exact same registered plugins as the general default
// registry - we just add this to add some specific agent registry methods.
var defaultRegistry = NewAgentRegistry(plugin.DefaultRegistry())

func NewAgentRegistry(parent *plugin.Registry) *AgentRegistry {
	r := &AgentRegistry{parent}

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
			// Skip non-agent plugins in registry, but still mark
			// as processed
			dataLock.Lock()
			errMap[name] = nil
			dataLock.Unlock()
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

func CollectHeartbeatData(ctx context.Context) (PluginHeartbeatDataList, map[plugin.PluginName]derrors.Error) {
	return defaultRegistry.CollectHeartbeatData(ctx)
}
