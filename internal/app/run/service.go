/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package run

// Agent normal operation

import (
	"fmt"
	"time"

	"github.com/nalej/derrors"
	"github.com/nalej/infra-net-plugin"

	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/config"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Service struct {
	Config *config.Config
	Client *client.AgentClient

	stopChan    chan struct{}
	disableChan chan struct{}
	lastBeat    time.Time
}

func (s *Service) Validate() derrors.Error {
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
func (s *Service) RestartPlugins() derrors.Error {
	plugins := s.Config.GetStringMap(plugin.DefaultPluginPrefix)
	for k, _ := range plugins {
		conf := s.Config.Sub(fmt.Sprintf("%s.%s", plugin.DefaultPluginPrefix, k))
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

func (s *Service) StartCorePlugin() derrors.Error {
	// We pass ourselves in the config, so that the core plugin can
	// control the service
	conf := viper.New()
	conf.Set("runner", s)
	conf.Set("config", s.Config)

	derr := plugin.StartPlugin("core", conf)
	if derr != nil {
		return derr
	}

	return nil
}

func (s *Service) Run() derrors.Error {
	if s.Client == nil {
		return derrors.NewInvalidArgumentError("client not set")
	}

	printRegisteredPlugins()
	s.Config.Print()

	derr := s.StartCorePlugin()
	if derr != nil {
		return derr
	}

	derr = s.RestartPlugins()
	if derr != nil {
		return derr
	}

	interval := s.Config.GetDuration("agent.interval")
	beatTimeout := interval / 2
	assetId := s.Config.GetString("agent.asset_id")

	log.Debug().Str("interval", interval.String()).Msg("running")

	// Create worker to execute operations
	worker := NewWorker(s.Config.GetSubConfig(plugin.DefaultPluginPrefix))

	// Create dispatcher for operations to workers
	dispatcher, derr := NewDispatcher(s.Client, worker, s.Config.GetInt("agent.opqueue_len"))
	if derr != nil {
		return derr
	}

	beater := Beater{
		client:     s.Client,
		dispatcher: dispatcher,
		assetId:    assetId,
	}

	// Start main heartbeat ticker
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial heartbeat so the edge controller knows we're running right away
	_, derr = beater.Beat(beatTimeout)
	if derr != nil {
		return derr
	}

	s.stopChan = make(chan struct{})
	s.disableChan = make(chan struct{})
	for s.stopChan != nil && s.disableChan != nil {
		select {
		case <-ticker.C:
			// Send heartbeat
			ok, derr := beater.Beat(beatTimeout)
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
			s.stopChan = nil
		case <-s.disableChan:
			s.disableChan = nil
		}
	}

	derr = dispatcher.Stop(s.Config.GetDuration("agent.shutdown_timeout"))

	// Only disabled, still waiting for stop
	if s.stopChan != nil {
		<-s.stopChan
	}

	log.Debug().Msg("stopped")
	return derr
}

func (s *Service) errChanRun(errChan chan<- derrors.Error) {
	derr := s.Run()
	errChan <- derr
}

func (s *Service) Start(errChan chan<- derrors.Error) derrors.Error {
	go s.errChanRun(errChan)
	return nil
}

func (s *Service) Stop() {
	// Recover to avoid race conditions stopping and uninstalling
	// simultaneously, or running multiple uninstalls
	defer func() {
		if r := recover(); r != nil {
			log.Debug().Msg("already stopping")
		}
	}()

	close(s.stopChan)
}

func (s *Service) Alive() (bool, derrors.Error) {
	// If last successfull main loop run is longer than twice the heartbeat
	// interval ago, we are not alive
	if time.Since(s.lastBeat) > 2*s.Config.GetDuration("agent.interval") {
		return false, nil
	}
	return true, nil
}

func (s *Service) Disable() {
	// Recover to avoid race conditions stopping and uninstalling
	// simultaneously, or running multiple uninstalls
	defer func() {
		if r := recover(); r != nil {
			log.Debug().Msg("already disabled")
		}
	}()

	log.Info().Msg("disabling agent heartbeat and operations")

	close(s.disableChan)
}
