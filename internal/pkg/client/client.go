/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package client

// Create client connection to Edge Controller

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/nalej/derrors"
	"github.com/nalej/grpc-edge-controller-go"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ConnectionOptions struct {
	UseTLS bool
	CACert string
	Insecure bool
}

type AgentClient struct {
	grpc_edge_controller_go.AgentClient
	*grpc.ClientConn
}

func NewAgentClient(address string, opts *ConnectionOptions) (*AgentClient, derrors.Error) {
	var options []grpc.DialOption

	log.Debug().Str("address", address).Msg("creating connection")

	if opts.UseTLS {
		var pool *x509.CertPool
		var err error
		if opts.CACert != "" {
			pool = x509.NewCertPool()
			derr := addCert(pool, opts.CACert)
			if derr != nil {
				return nil, derr
			}
		} else {
			// Use system pool
			pool, err = x509.SystemCertPool()
			if err != nil {
				return nil, derrors.NewInternalError("unable to initialize system certificate pool", err)
			}
		}

		if opts.Insecure {
			log.Warn().Msg("creating insecure connection")
		}

		tlsConfig := &tls.Config{
			RootCAs: pool,
			ServerName: strings.Split(address, ":")[0],
			InsecureSkipVerify: opts.Insecure,
		}

		creds := credentials.NewTLS(tlsConfig)
		log.Debug().Interface("creds", creds.Info()).Msg("secure credentials")
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		log.Warn().Msg("creating unencrypted connection")
		options = append(options, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(address, options...)
	if err != nil {
		return nil, derrors.NewInternalError("unable to create client connection", err).WithParams(address)
	}

	client := grpc_edge_controller_go.NewAgentClient(conn)

	return &AgentClient{client, conn}, nil
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
