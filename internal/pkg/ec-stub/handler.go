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

package ec_stub

// Handler for Edge Controller stub

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-edge-controller-go"
	"github.com/nalej/grpc-inventory-manager-go"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	opID uint64

	checksReceived    uint64
	callbacksReceived uint64
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) AgentJoin(ctx context.Context, request *grpc_edge_controller_go.AgentJoinRequest) (*grpc_inventory_manager_go.AgentJoinResponse, error) {
	log.Info().Interface("request", request).Msg("join request received")
	response := &grpc_inventory_manager_go.AgentJoinResponse{
		OrganizationId: "test-org",
		AssetId:        "test-asset",
		Token:          "test-token",
	}

	return response, nil
}

func (h *Handler) AgentCheck(ctx context.Context, request *grpc_edge_controller_go.AgentCheckRequest) (*grpc_edge_controller_go.CheckResult, error) {
	log.Info().Interface("request", request).Msg("heartbeat received")
	atomic.AddUint64(&h.checksReceived, 1)
	response := &grpc_edge_controller_go.CheckResult{
		PendingRequests: []*grpc_inventory_manager_go.AgentOpRequest{
			&grpc_inventory_manager_go.AgentOpRequest{
				AssetId:     "test-asset",
				Plugin:      "ping",
				Operation:   "start",
				OperationId: h.nextOpID(),
				Params: map[string]string{
					"test-key": "test-val",
				},
			},
			&grpc_inventory_manager_go.AgentOpRequest{
				AssetId:     "test-asset",
				Plugin:      "ping",
				Operation:   "ping",
				OperationId: h.nextOpID(),
			},
			&grpc_inventory_manager_go.AgentOpRequest{
				AssetId:     "test-asset",
				Plugin:      "ping",
				Operation:   "stop",
				OperationId: h.nextOpID(),
			},
		},
	}

	return response, nil
}

func (h *Handler) CallbackAgentOperation(ctx context.Context, request *grpc_inventory_manager_go.AgentOpResponse) (*grpc_common_go.Success, error) {
	log.Info().Interface("request", request).Msg("operation callback received")
	atomic.AddUint64(&h.callbacksReceived, 1)
	response := &grpc_common_go.Success{}
	return response, nil
}

func (h *Handler) nextOpID() string {
	return fmt.Sprintf("%d", atomic.AddUint64(&h.opID, 1))
}

func (h *Handler) GetNumChecks() uint64 {
	return atomic.LoadUint64(&h.checksReceived)
}

func (h *Handler) GetNumCallbacks() uint64 {
	return atomic.LoadUint64(&h.callbacksReceived)
}
