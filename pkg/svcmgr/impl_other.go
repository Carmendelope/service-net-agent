// +build !linux

/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Service manager - implementation for unsupported systems

package svcmgr

import (
	"go/build"

	"github.com/nalej/derrors"
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

// On unsupported systems we never run as service
func runningAsService() bool {
	return false
}

// On unsupported systems, the system services are never OK
func checkSystem() bool {
	return false
}
