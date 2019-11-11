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

// Service manager - interface to systemd and Windows API

package svcmgr

import (
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
)

type Runner interface {
	// Run only returns with an error - otherwise keeps running forever
	Run() derrors.Error

	// Start keeps running until an error occurs
	Start(errChan chan<- derrors.Error) derrors.Error
	Stop()

	// Disable all operations, but don't exit yet
	Disable()

	Alive() (bool, derrors.Error)
}

type Manager struct {
	Name   string
	runner Runner

	*Implementation
}

func NewManager(name string, runner Runner) (*Manager, derrors.Error) {
	impl, derr := NewImplementation(name, runner)
	if derr != nil {
		return nil, derr
	}

	m := &Manager{
		Name:           name,
		runner:         runner,
		Implementation: impl,
	}

	return m, nil
}

// Start
func (m *Manager) Run(asService bool) derrors.Error {
	// If we're not a service, no management is needed
	if !asService {
		return m.runner.Run()
	}

	derr := checkSystem()
	if derr != nil {
		return derr
	}

	log.Info().Str("name", m.Name).Msg("starting service")

	derr = m.Implementation.Run()
	if derr != nil {
		return derrors.NewInternalError("error running service", derr)
	}

	log.Info().Str("name", m.Name).Msg("service stopped")
	return nil
}
