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

// Edge Controller server stub

import (
	"net"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Start(grpcListener net.Listener, errChan chan<- error) (*grpc.Server, derrors.Error) {
	// Create server and register handler
	handler := NewHandler()
	grpcServer := grpc.NewServer()
	grpc_edge_controller_go.RegisterAgentServer(grpcServer, handler)

	// Start gRPC server
	reflection.Register(grpcServer)
	log.Info().Msg("Launching gRPC server")
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Error().Err(err).Msg("failed to serve grpc")
			errChan <- err
		}
		log.Info().Msg("closed grpc server")
	}()

	return grpcServer, nil
}
