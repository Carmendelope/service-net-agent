/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package ping

// Ping example plugin return data

import (
	"github.com/nalej/grpc-edge-controller-go"
)

type PingData struct {}

func (d *PingData) ToGRPC() *grpc_edge_controller_go.PluginData {
	data := &grpc_edge_controller_go.PluginData{
		// TODO: Make PING plugin data structures in gRPC for testing
	}

	return data
}
