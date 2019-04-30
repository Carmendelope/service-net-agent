/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package client

// Create client connection to Edge Controller from Config object

import (
	"github.com/nalej/derrors"
	"github.com/nalej/service-net-agent/internal/pkg/config"
)

func FromConfig(config *config.Config, token ...string) (*AgentClient, derrors.Error) {
	// Get token from config
	t := config.GetString("agent.token")
	// Override from optional argument
	if len(token) > 0 {
		t = token[0]
	}

	// Create connection
	opts := &ConnectionOptions{
		UseTLS: config.GetBool("controller.tls"),
		CACert: config.GetString("controller.cert"),
		Insecure: config.GetBool("controller.insecure"),
		Token: t,
	}

	return NewAgentClient(config.GetString("controller.address"), opts)
}
