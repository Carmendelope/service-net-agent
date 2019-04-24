// +build !linux

/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package inventory

// Gather inventory information

import (
	"go/build"

	"github.com/nalej/derrors"
	"github.com/nalej/grpc-edge-controller-go"

	"github.com/rs/zerolog/log"
)

// TBD - this package should not be OS-specific; that should be encapsulated
// in nalej/sysinfo. We'll tackle that when we clean that package up and
// introduce Windows compatibility.

type Inventory struct {
}

func NewInventory() (*Inventory, derrors.Error) {
	log.Warn().Str("os", build.Default.GOOS).Msg("gathering inventory not supported")
	i := &Inventory{
	}

	return i, nil
}

func (i *Inventory) GetRequest() (*grpc_edge_controller_go.AgentJoinRequest) {
	request := &grpc_edge_controller_go.AgentJoinRequest{
		Labels: map[string]string{},
	}
	return request
}
