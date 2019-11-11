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
