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

	err = copyFile(bin, destDir)
	if err != nil {
		return derrors.NewPermissionDeniedError("unable to copy file", err).WithParams(bin, destDir)
	}

	// Install system services
	// TBD

	return nil
}
