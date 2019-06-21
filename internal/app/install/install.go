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

type InstallCommandType string
const (
	InstallCommand InstallCommandType = "install"
	UninstallCommand InstallCommandType = "uninstall"
)

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

func (i *Installer) Run(command InstallCommandType) (derrors.Error) {
	i.Print()

	switch command {
	case InstallCommand:
		return i.Install()
	case UninstallCommand:
		return i.Uninstall()
	default:
		return derrors.NewInvalidArgumentError("invalid install command").WithParams(command)
	}
}

func (i *Installer) Install() (derrors.Error) {
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
	svcInstaller, derr := svcmgr.NewInstaller(defaults.AgentName, i.Config.Path)
	if derr != nil {
		return derr
	}

	args := []string{
		"run",
		"--service",
		"--config",
		i.Config.ConfigFile,
	}

	// Additional arguments based on configuration
	args = append(args, extraArgs(i.Config)...)

	derr = svcInstaller.Install(dest, args, defaults.AgentDescription)
	if derr != nil {
		return derr
	}

	derr = svcInstaller.Enable()
	if derr != nil {
		return derr
	}

	return nil
}

// Uninstall command will:
// - Stop the system service
// - Disable and remove the system service
// - Delete the configuration file
func (i *Installer) Uninstall() (derrors.Error) {
	return derrors.NewUnimplementedError("uninstall not implemented")
}
