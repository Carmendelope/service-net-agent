/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Agent normal operation

import (
	"fmt"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/rs/zerolog/log"
)

const pluginConfigKey = "plugins"

type Service struct {
	Config *config.Config
	Client *client.AgentClient

	stopChan chan bool
	lastBeat time.Time
}

func (s *Service) Validate() (derrors.Error) {
	if s.Config.GetString("agent.token") == "" {
		return derrors.NewFailedPreconditionError("no token found - agent not joined to edge controller")
	}
	if s.Config.GetString("agent.asset_id") == "" {
		return derrors.NewFailedPreconditionError("no asset id found - agent not joined to edge controller")
	}
	if s.Config.GetString("controller.address") == "" {
		return derrors.NewInvalidArgumentError("address must be specified")
	}
	if s.Config.GetDuration("agent.interval") <= 0 {
		return derrors.NewInvalidArgumentError("valid interval (> 0) must be specified")
	}

	return nil
}

// Restart previously running plugins
func (s *Service) RestartPlugins() (derrors.Error) {
	plugins := s.Config.GetStringMap(pluginConfigKey)
	for k, _ := range(plugins) {
		conf := s.Config.Sub(fmt.Sprintf("%s.%s", pluginConfigKey, k))
		if !conf.GetBool("enabled") {
			continue
		}
		derr := plugin.StartPlugin(plugin.PluginName(k), conf)
		if derr != nil {
			return derr
		}
	}

	return nil
}

func (s *Service) Run() (derrors.Error) {
	if s.Client == nil {
		return derrors.NewInvalidArgumentError("client not set")
	}

	printRegisteredPlugins()
	s.Config.Print()

	derr := s.RestartPlugins()
	if derr != nil {
		return derr
	}

	interval := s.Config.GetDuration("agent.interval")
	assetId := s.Config.GetString("agent.asset_id")

	log.Debug().Str("interval", interval.String()).Msg("running")

	// Create worker to execute operations
	worker := NewWorker(s.Config.GetSubConfig(pluginConfigKey))

	// Create dispatcher for operations to workers
	dispatcher, derr := NewDispatcher(s.Client, worker, s.Config.GetInt("agent.opqueue_len"))
	if derr != nil {
		return derr
	}

	// Start main heartbeat ticker
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial heartbeat so the edge controller knows we're running right away
	_, derr = heartbeat(s.Client, dispatcher, assetId)
	if derr != nil {
		return derr
	}

	stop := false
	for !stop {
		select {
		case <-ticker.C:
			// Send heartbeat
			ok, derr := heartbeat(s.Client, dispatcher, assetId)
			if derr != nil {
				// Something is wrong with the dispatcher,
				// we're not going to try to stop it.
				return derr
			}

			// Record last succesfull run
			if ok {
				s.lastBeat = time.Now()
			}
		case <-s.stopChan:
			stop = true
		}
	}

	derr = dispatcher.Stop(s.Config.GetDuration("agent.shutdown_timeout"))

	log.Debug().Msg("stopped")
	return derr
}

func (s *Service) errChanRun(errChan chan<- derrors.Error) {
	derr := s.Run()
	errChan <- derr
}

func (s *Service) Start(errChan chan<- derrors.Error) derrors.Error {
	s.stopChan = make(chan bool, 1)

	go s.errChanRun(errChan)
	return nil
}

func (s *Service) Stop() {
	s.stopChan <- true
}

func (s *Service) Alive() (bool, derrors.Error) {
	// If last successfull main loop run is longer than twice the heartbeat
	// interval ago, we are not alive
	if time.Since(s.lastBeat) > 2 * s.Config.GetDuration("agent.interval") {
		return false, nil
	}
	return true, nil
}

func heartbeat(client *client.AgentClient, dispatcher *Dispatcher, assetId string) (bool, derrors.Error) {
	log.Debug().Msg("sending heartbeat to edge controller")

	var beatSent = false

	// Create default heartbeat message
	beatRequest := &grpc_edge_controller_go.AgentCheckRequest{
		AssetId: assetId,
		Timestamp: time.Now().UTC().Unix(),
	}

	ctx := client.GetContext()
	result, err := client.AgentCheck(ctx, beatRequest)
	if err != nil {
		log.Warn().Err(err).Msg("failed sending heartbeat")
		return beatSent, nil
	}
	beatSent = true

	operations := result.GetPendingRequests()
	for _, operation := range(operations) {
		// Check asset id
		if operation.GetAssetId() != assetId {
			log.Warn().Str("operation_id", operation.GetOperationId()).
				Str("asset_id", operation.GetAssetId()).
				Msg("received operation with non-matching asset id")
			continue
		}

		derr := dispatcher.Dispatch(operation)
		if derr != nil {
			// Little risky to bail out of main loop when this fails,
			// but the scheduling of an operation really shouldn't
			// fail unless something is actually broken. We have a
			// watchdog that will restart the agent in such case.
			return beatSent, derr
		}
	}

	return beatSent, nil
}
