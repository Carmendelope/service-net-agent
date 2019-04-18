/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package join

// Join Agent to Nalej Edge

import (
	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
)

type Joiner struct {
	Config *config.Config

	Token string
}

func (j *Joiner) Validate() (derrors.Error) {
	if j.Config.GetString("controller.address") == "" {
		return derrors.NewInvalidArgumentError("address must be specified")
	}
	if j.Token == "" {
		return derrors.NewInvalidArgumentError("token must be specified")
	}

	return nil
}

func (j *Joiner) Run() (derrors.Error) {
	j.Config.Print()

	return nil
}
