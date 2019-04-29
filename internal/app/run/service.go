/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Agent normal operation

import (
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-inventory-go"

	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/config"

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
			if len(operations) > 0 {
				// TBD - Handle operation requests async (spawn goroutine)
				log.Debug().Interface("operations", operations).Msg("received operation requests")
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
