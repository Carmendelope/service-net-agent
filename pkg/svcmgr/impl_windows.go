/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Service manager - implementation for unsupported systems

package svcmgr

import (
	"go/build"

	"github.com/nalej/derrors"

	"golang.org/x/sys/windows/svc/mgr"
)

type Implementation struct {
}

func NewImplementation(Runner) (*Implementation, derrors.Error) {
	i := &Implementation{
	}

	return i, nil
}

func (i *Implementation) Run() derrors.Error {
	return derrors.NewUnimplementedError("starting as service not supported").WithParams(build.Default.GOOS)
}

// If we're running on Windows, we assume we can actually run services
func checkSystem() derrors.Error {
	return nil
}

func Start(servicename string) derrors.Error {
	// Connect to Windows service manager
	m, err := mgr.Connect()
	if err != nil {
		return derrors.NewInternalError("unable to connect to system service manager", err)
	}
	defer m.Disconnect()

	// Get service
	s, err := m.OpenService(servicename)
	if err != nil {
		return derrors.NewInvalidArgumentError("service not installed", err).WithParams(servicename)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return derrors.NewInvalidArgumentError("unable to start system service", err).WithParams(servicename)
	}

	return nil
}
