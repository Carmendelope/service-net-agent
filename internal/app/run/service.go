/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Agent normal operation

import (
	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
)

type Service struct {
	Config *config.Config
}

func (s *Service) Validate() (derrors.Error) {
	if s.Config.GetString("controller.address") == "" {
		return derrors.NewInvalidArgumentError("address must be specified")
	}
	if s.Config.GetInt("agent.interval") <= 0 {
		return derrors.NewInvalidArgumentError("valid interval (> 0) must be specified")
	}

	return nil
}

func (s *Service) Run() (derrors.Error) {
	s.Config.Print()

	return nil
}
