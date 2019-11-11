/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
	out := make([]*grpc_edge_controller_go.PluginData, 0, len(d))
	for _, data := range d {
		out = append(out, data.ToGRPC())
	}

	return out
}
