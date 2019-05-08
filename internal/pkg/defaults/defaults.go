/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package defaults

// Agent defaults

import (
	"os"
)

const (
	AgentName = "service-net-agent"
	AgentDescription = "Nalej Service Net Agent"

	AgentHeartbeatInterval = 30
	AgentCommTimeout = 15
	AgentShutdownTimeout = 60
	AgentOpQueueLen = 32

	// Used to generate a unique but safe agent id
	ApplicationID = "allyourbasearebelongtonalej"

	ConfigFile string = "etc" + string(os.PathSeparator) + "agent.yaml"
	LogFile string = "log" + string(os.PathSeparator) + "agent.log"
	BinDir string = "bin"
)
