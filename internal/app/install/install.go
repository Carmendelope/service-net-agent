/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package install

// Agent installer

import (
	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
)

type Installer struct {
	Config *config.Config
}

func (i *Installer) Run() (derrors.Error) {
	return nil
}
