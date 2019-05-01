/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
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
func (i *Installer) Install(bin string, args []string, desc ...string) (derrors.Error) {
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
		return derrors.NewInvalidArgumentError("service already installed").WithParams(i.name)
	}

	// Create the service and associated event log entries
	description := strings.Join(desc, " ")
	conf := mgr.Config{
		StartType: mgr.StartManual,
		ErrorControl: mgr.ErrorNormal,
		DisplayName: description,
		Description: description,
	}
	s, err = m.CreateService(i.name, fullPath, conf, args...)
	if err != nil {
		return derrors.NewInternalError("unable to create service", err).WithParams(i.name)
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(i.name, eventlog.Error | eventlog.Warning | eventlog.Info)
	if err != nil {
		s.Delete()
		return derrors.NewInternalError("unable to set up event log for service", err).WithParams(i.name)
	}

	return nil
}

// Enable system service
func (i *Installer) Enable() (derrors.Error) {
	log.Debug().Str("name", i.name).Msg("enabling system service")

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
	c.StartType = mgr.StartAutomatic

	// Update config
	err = s.UpdateConfig(c)
	if err != nil {
		return derrors.NewInternalError("unable to set new service config", err).WithParams(i.name)
	}

	return nil
}

/*
func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}
*/
