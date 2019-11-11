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

package run

// Agent normal operation

import (
	"context"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/nalej/service-net-agent/internal/pkg/agentplugin"
	"github.com/nalej/service-net-agent/internal/pkg/client"

	"github.com/rs/zerolog/log"
)

type Beater struct {
	client     *client.AgentClient
	dispatcher *Dispatcher
	assetId    string
}

func (b *Beater) Beat(timeout time.Duration) (bool, derrors.Error) {
	log.Debug().Msg("sending heartbeat to edge controller")

	var beatSent = false

	beatCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	beatData, beatErrs := agentplugin.CollectHeartbeatData(beatCtx)

	// Warn about errors
	// TODO: Decide if we're not alive anymore due to errors
	for name, derr := range beatErrs {
		log.Warn().Err(derr).Str("plugin", name.String()).Msg("plugin error")
		log.Debug().Str("trace", derr.DebugReport()).Str("plugin", name.String()).Msg("plugin error trace")
	}

	// Create default heartbeat message
	beatRequest := &grpc_edge_controller_go.AgentCheckRequest{
		AssetId:    b.assetId,
		Timestamp:  time.Now().UTC().Unix(),
		PluginData: beatData.ToGRPC(),
	}

	ctx := b.client.GetContext()
	result, err := b.client.AgentCheck(ctx, beatRequest)
	if err != nil {
		log.Warn().Err(err).Msg("failed sending heartbeat")
		return beatSent, nil
	}
	beatSent = true

	operations := result.GetPendingRequests()
	for _, operation := range operations {
		// Check asset id
		if operation.GetAssetId() != b.assetId {
			log.Warn().Str("operation_id", operation.GetOperationId()).
				Str("asset_id", operation.GetAssetId()).
				Msg("received operation with non-matching asset id")
			continue
		}

		derr := b.dispatcher.Dispatch(operation)
		if derr != nil {
			// Little risky to bail out of main loop when this fails,
			// but the scheduling of an operation really shouldn't
			// fail unless something is actually broken. We have a
			// watchdog that will restart the agent in such case.
			return beatSent, derr
		}
	}

	return beatSent, nil
}
