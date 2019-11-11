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

// Service manager - implementation for unsupported systems

package svcmgr

import (
	"time"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/mgr"
)

type Implementation struct {
	name   string
	runner Runner
}

func NewImplementation(name string, runner Runner) (*Implementation, derrors.Error) {
	i := &Implementation{
		name:   name,
		runner: runner,
	}

	return i, nil
}

func (i *Implementation) Run() derrors.Error {
	// If we're running as a service, we hook into the Windows service
	// manager. If not, we use a debug component to run the same code
	// from the command line.
	run := svc.Run // Real service

	// Check if started as service
	interactive, err := svc.IsAnInteractiveSession()
	if err != nil {
		return derrors.NewInternalError("unable to determine if this is an interactive session", err)
	}
	if interactive {
		log.Info().Msg("starting interactive service")
		run = debug.Run // Command-line fake service
	}

	err = run(i.name, i) // We implement Execute so we're a valid service handler
	if err != nil {
		log.Error().Err(err).Msg("error running system service")
		return derrors.NewInternalError("service failed", err).WithParams(i.name)
	}

	return nil
}

func (i *Implementation) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	errChan := make(chan derrors.Error, 1)
	derr := i.runner.Start(errChan)
	if derr != nil {
		log.Error().Err(derr).Str("deub", derr.DebugReport()).Msg("unable to start service")
		errno = uint32(windows.ERROR_EXCEPTION_IN_SERVICE)
		return
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for {
		select {
		case err := <-errChan:
			// If we stop, errChan receives nil
			if err != nil {
				log.Error().Err(err).Msg("service returned error")
				errno = uint32(windows.ERROR_EXCEPTION_IN_SERVICE)
				// TBD: Service-specific error codes - not sure how useful
			}

			changes <- svc.Status{State: svc.Stopped}
			return
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				// We don't independently change status or the
				// commands we accept
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Info().Msg("Gracefully shutting down")
				changes <- svc.Status{State: svc.StopPending}
				i.runner.Stop()
			default:
				log.Error().Interface("request", c).Msg("unexpected control request")
				errno = uint32(windows.ERROR_INVALID_SERVICE_CONTROL)
				return
			}
		}
	}

	return
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

	log.Info().Str("name", servicename).Msg("service started")
	return nil
}

// Stop system service
func Stop(servicename string) derrors.Error {
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

	status, err := s.Control(svc.Stop)
	if err != nil {
		return derrors.NewInvalidArgumentError("unable to stop system service", err).WithParams(servicename)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped {
		if timeout.Before(time.Now()) {
			return derrors.NewInternalError("timeout waiting for service to stop")
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return derrors.NewInternalError("could not retrieve service status", err)
		}
	}
	log.Info().Str("name", servicename).Msg("service stopped")
	return nil
}
