/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Agent normal operation

import (
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-inventory-manager-go"
	"github.com/nalej/grpc-inventory-go"

	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/rs/zerolog/log"
)

type Service struct {
	Config *config.Config

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

func (s *Service) Run() (derrors.Error) {
	s.Config.Print()

	interval := s.Config.GetDuration("agent.interval")
	assetId := s.Config.GetString("agent.asset_id")

	// Create client connection
	client, derr := client.FromConfig(s.Config)
	if derr != nil {
		return derr
	}
	ctx := client.GetContext()

	log.Debug().Str("interval", interval.String()).Msg("running")

	// Send start message
	request := &grpc_inventory_manager_go.AgentStartInfo{
		AssetId: assetId,
		Ip: client.LocalAddress(),
	}
	log.Debug().Str("asset_id", request.AssetId).Str("local_ip", request.Ip).Msg("sending start message to edge controller")
	_, err := client.AgentStart(ctx, request)
	if err != nil {
		return derrors.NewUnavailableError("unable to send start message to edge controller", err)
	}

	// Create default heartbeat message
	beatRequest := &grpc_inventory_go.AssetId{
		AssetId: assetId,
	}

	// Start main heartbeat ticker
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	stop := false
	for !stop {
		select {
		case <-ticker.C:
			// Send heartbeat
			log.Debug().Msg("sending heartbeat to edge controller")
			result, err := client.AgentCheck(ctx, beatRequest)
			if err != nil {
				log.Warn().Err(err).Msg("failed sending heartbeat")
				continue
			}

			operations := result.GetPendingRequests()
			for _, operation := range(operations) {
				derr := s.ScheduleOperation(operation)
				if derr != nil {
					// Little risky to bail out of main loop when this fails,
					// but the scheduling of an operation really shouldn't
					// fail unless something is actually broken. We have a
					// watchdog that will restart the agent in such case.
					return derr
				}
			}

			// Record last succesfull run
			s.lastBeat = time.Now()
		case <-s.stopChan:
			stop = true
		}
	}

	log.Debug().Msg("stopped")
	return nil
}

func (s *Service) ScheduleOperation(request *grpc_inventory_manager_go.AgentOpRequest) derrors.Error {
	var pluginName = plugin.PluginName(request.GetPlugin())
	var opName = plugin.CommandName(request.GetOperation())
	var params = request.GetParams()

	log.Debug().Str("plugin", pluginName.String()).Str("operation", opName.String()).Interface("params", params).Msg("received operation request")

	// TBD
	// Job queue with max capacity
	// Answer scheduled of queued, failed with specific error if capacity reached
	// only a single worker - we need to execute in order and we don't want to overload agent
	// timeout on operations

	// Add to request list
	// have executor thread
	return nil
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
