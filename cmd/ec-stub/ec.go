/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/ec-stub"

	"github.com/rs/zerolog/log"
)

func main() {
	derr := Run()
	if derr != nil {
		log.Fatal().Err(derr).Str("debug", derr.DebugReport()).Msg("failed")
	}
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

	grpcServer, derr := ec_stub.Start(grpcListener, errChan)
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
