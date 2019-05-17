/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package ec_stub

// Edge Controller server stub

import (
	"net"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/rs/zerolog/log"
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
