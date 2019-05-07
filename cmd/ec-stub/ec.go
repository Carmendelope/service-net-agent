/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-edge-controller-go"
	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/grpc-inventory-manager-go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/rs/zerolog/log"
)

func main() {
	derr := Run()
	if derr != nil {
		log.Fatal().Err(derr).Str("debug", derr.DebugReport()).Msg("failed")
	}
}

var opID = 0
// Not thread safe
func getopid() string {
	opID += 1
	return fmt.Sprintf("%d", opID)
}

// Run the service, launch the REST service handler.
func Run() derrors.Error {
	// Channel to signal errors from starting the servers
	errChan := make(chan error, 1)

	// Start listening on API port
	grpcListener, err := net.Listen("tcp", ":12345")
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	grpcServer, derr := start(grpcListener, errChan)
	if derr != nil {
		return derr
	}
	defer grpcServer.GracefulStop()

	// Wait for termination signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)

	select {
	case sig := <-sigterm:
		log.Info().Str("signal", sig.String()).Msg("Gracefully shutting down")
	case err := <-errChan:
		// We've already logged the error
		return derrors.NewInternalError("failed starting server", err)
	}

	return nil
}

func start(grpcListener net.Listener, errChan chan<- error) (*grpc.Server, derrors.Error) {
	// Create server and register handler
	handler := &Handler{}
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

type Handler struct {

}

func (h *Handler) AgentJoin(ctx context.Context, request *grpc_edge_controller_go.AgentJoinRequest) (*grpc_inventory_manager_go.AgentJoinResponse, error) {
	log.Info().Interface("request", request).Msg("join request received")
	response := &grpc_inventory_manager_go.AgentJoinResponse{
		OrganizationId: "test-org",
		AssetId: "test-asset",
		Token: "test-token",
	}

	return response, nil
}

func (h *Handler) AgentCheck(ctx context.Context, request *grpc_inventory_go.AssetId) (*grpc_edge_controller_go.CheckResult, error) {
	log.Info().Interface("request", request).Msg("heartbeat received")
	response := &grpc_edge_controller_go.CheckResult{
		PendingRequests: []*grpc_inventory_manager_go.AgentOpRequest{
			&grpc_inventory_manager_go.AgentOpRequest{
				AssetId: "test-asset",
				Plugin: "ping",
				Operation: "start",
				OperationId: getopid(),
				Params: map[string]string{
					"test-key": "test-val",
				},
			},
			&grpc_inventory_manager_go.AgentOpRequest{
				AssetId: "test-asset",
				Plugin: "ping",
				Operation: "ping",
				OperationId: getopid(),
			},
			&grpc_inventory_manager_go.AgentOpRequest{
				AssetId: "test-asset",
				Plugin: "ping",
				Operation: "stop",
				OperationId: getopid(),
			},

		},
	}

	return response, nil
}

func (h *Handler) CallbackAgentOperation(ctx context.Context, request *grpc_inventory_manager_go.AgentOpResponse) (*grpc_common_go.Success, error) {
	log.Info().Interface("request", request).Msg("operation callback received")
	response := &grpc_common_go.Success{}
	return response, nil
}

