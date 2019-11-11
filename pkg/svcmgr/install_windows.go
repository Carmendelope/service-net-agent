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

// Service manager - service installation for Windows

package svcmgr

import (
	"strings"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"

	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

type Installer struct {
	name string
	root string
}

func NewInstaller(name, root string) (*Installer, derrors.Error) {
	i := &Installer{
		name: name,
		root: root,
	}

	return i, nil
}

// Install a service
func (i *Installer) Install(bin string, args []string, desc ...string) derrors.Error {
	log.Debug().Str("name", i.name).Str("bin", bin).Msg("installing service")

	// Check if exists and executable
	fullPath, derr := fullExecPath(bin)
	if derr != nil {
		return derr
	}

	// Connect to Windows service manager
	m, err := mgr.Connect()
	if err != nil {
		return derrors.NewInternalError("unable to connect to system service manager", err)
	}
	defer m.Disconnect()

	// Check if service is already installed
	s, err := m.OpenService(i.name)
	if err == nil {
		s.Close()
		return derrors.NewInvalidArgumentError("service already installed, removing").WithParams(i.name)
	}

	// Create the service and associated event log entries
	description := strings.Join(desc, " ")
	conf := mgr.Config{
		StartType:    mgr.StartManual,
		ErrorControl: mgr.ErrorNormal,
		DisplayName:  description,
		Description:  description,
	}
	s, err = m.CreateService(i.name, fullPath, conf, args...)
	if err != nil {
		return derrors.NewInternalError("unable to create service", err).WithParams(i.name)
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(i.name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return derrors.NewInternalError("unable to set up event log for service", err).WithParams(i.name)
	}

	return nil
}

func (i *Installer) Remove() derrors.Error {
	log.Debug().Str("name", i.name).Msg("removing service")

	m, err := mgr.Connect()
	if err != nil {
		return derrors.NewInternalError("unable to connect to system service manager", err)
	}
	defer m.Disconnect()

	// Get service
	s, err := m.OpenService(i.name)
	if err != nil {
		return derrors.NewInvalidArgumentError("service not installed", err).WithParams(i.name)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return derrors.NewInternalError("unable to delete service", err).WithParams(i.name)
	}

	err = eventlog.Remove(i.name)
	if err != nil {
		return derrors.NewInternalError("unable to remove event log entries", err).WithParams(i.name)
	}

	// TODO: Reboot?

	return nil
}

// Enable system service
func (i *Installer) Enable() derrors.Error {
	log.Debug().Str("name", i.name).Msg("enabling system service")

	return i.setAutoManual(mgr.StartAutomatic)
}

// Disable system service
func (i *Installer) Disable() derrors.Error {
	log.Debug().Str("name", i.name).Msg("disabling system service")

	return i.setAutoManual(mgr.StartManual)
}

func (i *Installer) setAutoManual(startType uint32) derrors.Error {
	// Connect to Windows service manager
	m, err := mgr.Connect()
	if err != nil {
		return derrors.NewInternalError("unable to connect to system service manager", err)
	}
	defer m.Disconnect()

	// Get service
	s, err := m.OpenService(i.name)
	if err != nil {
		return derrors.NewInvalidArgumentError("service not installed", err).WithParams(i.name)
	}
	defer s.Close()

	// Get current config
	c, err := s.Config()
	if err != nil {
		return derrors.NewInternalError("unable to retrieve service config", err).WithParams(i.name)
	}

	// Set auto-start
	c.StartType = startType

	// Update config
	err = s.UpdateConfig(c)
	if err != nil {
		return derrors.NewInternalError("unable to set new service config", err).WithParams(i.name)
	}

	return nil
}
