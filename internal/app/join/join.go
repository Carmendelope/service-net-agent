/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package join

// Join Agent to Nalej Edge

import (
	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"
	"github.com/nalej/service-net-agent/internal/pkg/inventory"
	"github.com/nalej/service-net-agent/pkg/machine_id"

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

	// Get join request with inventory
	request, derr := j.getRequest()
	if derr != nil {
		return derr
	}

	// Create connection
	client, derr := client.FromConfig(j.Config, j.Token)
	if derr != nil {
		return derr
	}
	defer client.Close()

	// Send request and get agent token
	response, err := client.AgentJoin(client.GetContext(), request)
	if err != nil {
		return derrors.NewUnavailableError("unable to send join request", err)
	}

	// Store agent token in config
	j.Config.Set("agent.token", response.GetToken())
	j.Config.Set("agent.asset_id", response.GetAssetId())

	// Write config
	derr = j.Config.Write()
	if derr != nil {
		return derr
	}

	return nil
}

func (j *Joiner) getRequest() (*grpc_edge_controller_go.AgentJoinRequest, derrors.Error) {
	// Gather inventory
	inv, derr := inventory.NewInventory()
	if derr != nil {
		return nil, derr
	}

	request := inv.GetRequest()

	// Add command line-specified labels
	for k, v := range(j.Labels) {
		request.Labels[k] = v
	}

	// Add agent ID
	request.AgentId = machine_id.AppSpecificMachineID(defaults.ApplicationID)
	log.Info().Interface("inventory", request).Msg("system info")

	return request, nil
}
