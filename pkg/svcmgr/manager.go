/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
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

	Alive() (bool, derrors.Error)
}

type Manager struct {
	Name string
	runner Runner

	*Implementation
}

func NewManager(name string, runner Runner) (*Manager, derrors.Error) {
	impl, derr := NewImplementation(name, runner)
	if derr != nil {
		return nil, derr
	}

	m := &Manager{
		Name: name,
		runner: runner,
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
