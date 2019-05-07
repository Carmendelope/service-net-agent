/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package defaults

// Agent defaults

const (
	AgentName = "service-net-agent"
	AgentDescription = "Nalej Service Net Agent"

	AgentHeartbeatInterval = 30
	AgentCommTimeout = 15
	AgentShutdownTimeout = 60
	AgentOpQueueLen = 32

	// Used to generate a unique but safe agent id
	ApplicationID = "allyourbasearebelongtonalej"

	// Linux defaults
	Path string = "/opt/nalej"
	ConfigFile string = "etc/agent.yaml"
	BinDir string = "bin"
)
