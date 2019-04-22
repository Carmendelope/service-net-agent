// +build linux

/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package install

// Agent installer

import (
	"os"
	"path/filepath"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"
	"github.com/nalej/service-net-agent/pkg/svcmgr"
	"github.com/nalej/service-net-agent/version"

	"github.com/rs/zerolog/log"
)

type Installer struct {
	Config *config.Config
}

func (i *Installer) Validate() (derrors.Error) {
	if i.Config.Path == "" {
		return derrors.NewInvalidArgumentError("path must be specified")
	}

	return nil
}

func (i *Installer) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	log.Info().Str("path", i.Config.Path).Msg("installation path")
}

func (i *Installer) Run() (derrors.Error) {
	i.Print()

	// Create destination
	destDir := filepath.Join(i.Config.Path, defaults.BinDir)
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		return derrors.NewPermissionDeniedError("unable to create installation destination", err).WithParams(destDir)
	}

	// Copy myself
	bin, err := os.Executable()
	if err != nil {
		return derrors.NewPermissionDeniedError("unable to determine agent binary location", err)
	}

	// Destination is destination dir plus filename
	dest := filepath.Join(destDir, filepath.Base(bin))

	err = copyFile(dest, bin)
	if err != nil {
		return derrors.NewPermissionDeniedError("unable to copy file", err).WithParams(bin, destDir)
	}

	// Install system services
	svcInstaller, derr := svcmgr.NewInstaller(defaults.AgentName)
	if derr != nil {
		return derr
	}

	// TBD args - configpath

	derr = svcInstaller.Install(i.Config.Path, dest, []string{}, defaults.AgentDescription)
	if derr != nil {
		return derr
	}
	// TBD Enable

	return nil
}
