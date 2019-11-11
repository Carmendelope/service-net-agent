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

// Service manager - systemd implementation for linux

package svcmgr

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
)

type Implementation struct {
	runner Runner
}

func NewImplementation(name string, runner Runner) (*Implementation, derrors.Error) {
	i := &Implementation{
		runner: runner,
	}

	return i, nil
}

// Enable system service
func (i *Implementation) Run() derrors.Error {
	errChan := make(chan derrors.Error, 1)
	derr := i.runner.Start(errChan)
	if derr != nil {
		return derr
	}

	notifyReady()

	watchdogStopChan := make(chan bool, 1)
	go i.watchdog(watchdogStopChan)

	// Wait for termination signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)

	for {
		// We loop so we wait for errChan to receive nil (or error)
		// after calling Stop()
		select {
		case sig := <-sigterm:
			log.Info().Str("signal", sig.String()).Msg("Gracefully shutting down")
			watchdogStopChan <- true
			notifyStopping()
			i.runner.Stop()
		case err := <-errChan:
			if err != nil {
				log.Error().Err(err).Msg("service returned error")
				return derrors.NewInternalError("service returned error", err)
			}
			return nil
		}
	}
}

// Liveliness
func (i *Implementation) watchdog(stopChan <-chan bool) {
	interval, err := watchdogEnabled()
	if err != nil {
		log.Error().Err(err).Msg("failed retrieving system watchdog status")
		return
	}

	if interval == 0 {
		log.Info().Msg("system watchdog disabled")
		return
	}

	tickInterval := interval / 2
	log.Info().Str("watchdog_interval", interval.String()).Str("ticker_interval", tickInterval.String()).Msg("starting watchdog loop")

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			log.Info().Msg("stopping watchdog loop")
			return
		case <-ticker.C:
			log.Debug().Msg("watchdog checking liveliness")
			ok, derr := i.runner.Alive()
			if derr != nil {
				log.Error().Err(err).Msg("alive call failed")
				continue
			}
			if ok {
				log.Debug().Msg("main loop alive")
				notifyAlive()
			} else {
				log.Warn().Msg("main loop dead")
			}
		}
	}
}

func Start(servicename string) derrors.Error {
	// SystemD implementation
	return start(servicename)
}

// Stop system service
func Stop(servicename string) derrors.Error {
	log.Debug().Str("name", servicename).Msg("stopping system service")
	return stop(servicename)
}
