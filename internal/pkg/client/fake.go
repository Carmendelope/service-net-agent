/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package client

// Fake client for testing

import (
	"github.com/nalej/grpc-edge-controller-go"

	"google.golang.org/grpc"
)

func NewFakeAgentClient(conn *grpc.ClientConn) *AgentClient {
	return &AgentClient{
		AgentClient: grpc_edge_controller_go.NewAgentClient(conn),
		ClientConn: conn,
		opts: &ConnectionOptions{},
	}
}
