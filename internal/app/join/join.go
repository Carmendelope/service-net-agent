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

package join

// Join Agent to Nalej Edge

import (
	"fmt"
	"github.com/nalej/grpc-utils/pkg/conversions"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"
	"github.com/nalej/service-net-agent/internal/pkg/inventory"

	"github.com/denisbrodbeck/machineid"
	"github.com/rs/zerolog/log"
)

type Joiner struct {
	Config *config.Config

	Token  string
	Labels map[string]string
}

func (j *Joiner) Validate() derrors.Error {
	if j.Config.GetString("controller.address") == "" {
		return derrors.NewInvalidArgumentError("address must be specified")
	}
	if j.Token == "" {
		return derrors.NewInvalidArgumentError("token must be specified")
	}

	return nil
}

func (j *Joiner) Run() derrors.Error {
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
		log.Warn().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("unable to join")
		return derrors.NewUnavailableError("unable to send join request", err)
	}

	// Check and store join response (token and asset id)
	token := response.GetToken()
	assetId := response.GetAssetId()
	if token == "" {
		return derrors.NewInvalidArgumentError("no agent token received")
	}
	if assetId == "" {
		return derrors.NewInvalidArgumentError("no asset id received")
	}

	j.Config.Set("agent.token", token)
	j.Config.Set("agent.asset_id", assetId)

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
	for k, v := range j.Labels {
		request.Labels[k] = v
	}

	// Add agent ID
	id, err := machineid.ProtectedID(defaults.ApplicationID)
	if err != nil {
		return nil, derrors.NewInternalError("unable to retrieve machine id", err)
	}

	// Create dashed uuid format from application-specific machine id
	// This uses only the first 16 bytes of the ID instead of the full
	// 32, but that should still be unique.
	request.AgentId = fmt.Sprintf("%s-%s-%s-%s-%s", id[0:8], id[8:12], id[12:16], id[16:20], id[20:32])
	log.Info().Interface("inventory", request).Msg("system info")

	return request, nil
}
