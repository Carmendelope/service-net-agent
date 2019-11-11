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

package client

// Create client connection to Edge Controller

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/nalej/derrors"
	"github.com/nalej/grpc-edge-controller-go"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type ConnectionOptions struct {
	UseTLS   bool
	CACert   string
	Insecure bool
	Timeout  time.Duration
	Token    string
}

type AgentClient struct {
	grpc_edge_controller_go.AgentClient
	*grpc.ClientConn
	address string
	opts    *ConnectionOptions
}

func NewAgentClient(address string, opts *ConnectionOptions) (*AgentClient, derrors.Error) {
	log.Debug().Str("address", address).Msg("creating connection")

	agentClient := &AgentClient{
		address: address,
		opts:    opts,
	}

	dialOpts, derr := agentClient.getDialOptions()
	if derr != nil {
		return nil, derr
	}

	conn, err := grpc.Dial(address, dialOpts...)
	if err != nil {
		return nil, derrors.NewInternalError("unable to create client connection", err).WithParams(address)
	}

	agentClient.ClientConn = conn
	agentClient.AgentClient = grpc_edge_controller_go.NewAgentClient(conn)

	return agentClient, nil
}

func (c *AgentClient) GetContext() context.Context {
	meta := metadata.New(map[string]string{"Authorization": c.opts.Token})
	ctx := metadata.NewOutgoingContext(context.Background(), meta)
	if c.opts.Timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, c.opts.Timeout)
	}
	return ctx
}

// Get local address used for connecting to server
func (c *AgentClient) LocalAddress() string {
	// We cannot directly determine the local peer address from a gRPC
	// connection and we don't have access to the raw net.Conn. Hence,
	// we set up a dummy connection (we don't actually send any packets
	// when using a UDP connection) to figure/ out the local IP we would
	// be using.
	// See also the Golang source: srcAddrs() in net/addrselect.go.
	conn, err := net.Dial("udp", c.address)
	if err != nil {
		log.Warn().Err(err).Str("address", c.address).Msg("no route to server")
		return ""
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

// Get the dial options based on the ConnectionOptions
func (c *AgentClient) getDialOptions() ([]grpc.DialOption, derrors.Error) {
	var options []grpc.DialOption

	if c.opts.UseTLS {
		// A nil certificate pool for RootCAs in a tls.Config uses
		// the system certificates to validate servers, in a
		// cross-platform way.
		var pool *x509.CertPool = nil
		if c.opts.CACert != "" {
			pool = x509.NewCertPool()
			derr := addCert(pool, c.opts.CACert)
			if derr != nil {
				return nil, derr
			}
		}

		if c.opts.Insecure {
			log.Warn().Msg("creating insecure connection")
		}

		tlsConfig := &tls.Config{
			RootCAs:            pool,
			ServerName:         "", // we don't need to check the serverName
			InsecureSkipVerify: c.opts.Insecure,
		}

		creds := credentials.NewTLS(tlsConfig)
		log.Debug().Interface("creds", creds.Info()).Msg("secure credentials")

		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		log.Warn().Msg("creating unencrypted connection")
		options = append(options, grpc.WithInsecure())
	}

	return options, nil
}

// Add X509 certificate from a file to a pool
func addCert(pool *x509.CertPool, cert string) derrors.Error {
	caCert, err := ioutil.ReadFile(cert)
	if err != nil {
		return derrors.NewInternalError("unable to read certificate", err)
	}

	added := pool.AppendCertsFromPEM(caCert)
	if !added {
		return derrors.NewInternalError(fmt.Sprintf("Failed to add certificate from %s", cert))
	}

	return nil
}
