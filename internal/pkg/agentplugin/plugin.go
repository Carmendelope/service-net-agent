/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package agentplugin

// Agent plugin infrastructure

import (
	"context"

	"github.com/nalej/derrors"
	"github.com/nalej/grpc-edge-controller-go"
	"github.com/nalej/infra-net-plugin"
)

type AgentPlugin interface {
	plugin.Plugin

	// Callback for plugin to add data to heartbeat. Returns nil if
	// nothing needs to be added
	Beat(ctx context.Context) (PluginHeartbeatData, derrors.Error)
}

type PluginHeartbeatData interface {
	ToGRPC() *grpc_edge_controller_go.PluginData
}

type PluginHeartbeatDataList []PluginHeartbeatData
func (d PluginHeartbeatDataList) ToGRPC() []*grpc_edge_controller_go.PluginData {
	out := make ([]*grpc_edge_controller_go.PluginData, 0, len(d))
	for _, data := range(d) {
		out = append(out, data.ToGRPC())
	}

	return out
}
