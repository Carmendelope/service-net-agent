/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package install

// Agent installer

import (
	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
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

	return derrors.NewUnimplementedError("install not implemented")
}
