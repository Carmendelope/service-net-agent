/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package join

// Join Agent to Nalej Edge

import (
	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/inventory"

	"github.com/rs/zerolog/log"
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

	// Gather inventory
	inv, derr := inventory.NewInventory()
	if derr != nil {
		return derr
	}
	log.Info().Interface("inventory", inv).Msg("system info")

	// Join EC - get agent token
	request := &grpc_edge_controller_go.AgentJoinRequest{
		Os: inv.Os,
		Hardware: inv.Hardware,
		Storage: inv.Storage,
	}
	_ = request
	// TBD - create client, send request

	// Write config - store agent token in config
	j.Config.Set("agent.token", "TBD") // Replace with token from join response
	derr = j.Config.Write()
	if derr != nil {
		return derr
	}

	return nil
}
