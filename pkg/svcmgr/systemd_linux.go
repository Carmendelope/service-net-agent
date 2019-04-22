/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Service manager - interface to systemd and Windows API

package svcmgr

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
	"github.com/coreos/go-systemd/util"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
)

const (
	// Relative unit path
	systemDUnitPath = "systemd/system"
	systemDUnitExt = "service"

	sectionUnit = "Unit"
	sectionService = "Service"
	sectionInstall = "Install"

	nameDescription = "Description"
	nameAfter = "After"
	nameType = "Type"
	nameNotifyAccess = "NotifyAccess"
	nameRestart = "Restart"
	nameExecStart = "ExecStart"
	nameWantedBy = "WantedBy"

	valueNetwork = "network.target"
	valueNotify = "notify"
	valueMain = "main"
	valueOnFailure = "on-failure"
	valueMultiUser = "multi-user.target"
)

// Determine the absolute path for the unit file
func getUnitFilename(name, basePath string) (string, derrors.Error) {
	filename := fmt.Sprintf("%s.%s", name, systemDUnitExt)
        path := filepath.Join(basePath, systemDUnitPath, filename)
	return path, nil
}

// Create unit file
func createUnitFile(desc, bin string, args []string) (io.Reader, derrors.Error) {
	execCommand := fmt.Sprintf("%s %s", bin, strings.Join(args, " "))

	unitOpts := []*unit.UnitOption{
		unit.NewUnitOption(sectionUnit, nameAfter, valueNetwork),

		unit.NewUnitOption(sectionService, nameType, valueNotify),
		unit.NewUnitOption(sectionService, nameNotifyAccess, valueMain),
		unit.NewUnitOption(sectionService, nameRestart, valueOnFailure),
		unit.NewUnitOption(sectionService, nameExecStart, execCommand),

		unit.NewUnitOption(sectionInstall, nameWantedBy, valueMultiUser),
	}

	if desc != "" {
		unitOpts = append(unitOpts, unit.NewUnitOption(sectionUnit, nameDescription, desc))
	}

	// Create Unit file
	reader := unit.Serialize(unitOpts)

	return reader, nil
}

func enableUnit(unit string) derrors.Error {
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		return derrors.NewInternalError("unable to connect to system service manager", err)
	}
	defer conn.Close()

	_, status, err := conn.EnableUnitFiles([]string{unit}, false /* not runtime-only */, true /* force */)
	if err != nil {
		return derrors.NewInternalError("unable to enable system service", err)
	}

	for _, s := range(status) {
		log.Debug().Str("type", s.Type).Str("link", s.Filename).Str("unit", s.Destination).Msg("service status changed")
	}

	return nil
}

func notify(state string) {
	ok, err := daemon.SdNotify(false, state)
	if err != nil {
		log.Warn().Err(err).Str("state", state).Msg("failed notifying system of state")
		return
	}
	if !ok {
		log.Warn().Str("state", state).Msg("system notification not supported")
		return
	}

	log.Debug().Str("state", state).Msg("system notification sent")
}

func notifyReady() {
	notify(daemon.SdNotifyReady)
}

func notifyStopping() {
	notify(daemon.SdNotifyStopping)
}

func notifyAlive() {
	notify(daemon.SdNotifyWatchdog)
}

func watchdogEnabled() (time.Duration, error) {
	return daemon.SdWatchdogEnabled(false)

}

func checkSystem() derrors.Error {
	if !util.IsRunningSystemd() {
		return derrors.NewFailedPreconditionError("system is not running systemd")
	}
	return nil
}

func Start(servicename string) derrors.Error {
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		return derrors.NewInternalError("unable to connect to system service manager", err)
	}
	defer conn.Close()

	msgChan := make(chan string, 1)
	unit := fmt.Sprintf("%s.%s", servicename, systemDUnitExt)
	_, err = conn.StartUnit(unit, "replace", msgChan)
	if err != nil {
		return derrors.NewInternalError("failed to request start from system service manager", err).WithParams(unit)
	}

	msg := <-msgChan
	switch msg {
	case "done":
		log.Info().Msg("service started")
	case "canceled", "timeout", "failed":
		log.Error().Str("reason", msg).Msg("start failed")
	case "dependency":
		log.Error().Msg("service not started because of a failed dependency")
	case "skipped":
		log.Warn().Msg("service skipped; service not applicable to currently running units")
	}

	return nil
}
