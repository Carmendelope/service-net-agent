/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package join

// Join Agent to Nalej Edge

import (
	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/inventory"

	"github.com/rs/zerolog/log"
)

type Joiner struct {
	Config *config.Config

	Token string
	Labels map[string]string
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

	// Gather inventory
	inv, derr := inventory.NewInventory()
	if derr != nil {
		return derr
	}

	request := inv.GetRequest()

	// Add command line-specified labels
	for k, v := range(j.Labels) {
		request.Labels[k] = v
	}

	log.Info().Interface("inventory", request).Msg("system info")

	// Join EC - get agent token
	// TBD - create client, send request

	// Write config - store agent token in config
	j.Config.Set("agent.token", "TBD") // Replace with token from join response
	derr = j.Config.Write()
	if derr != nil {
		return derr
	}

	return nil
}
