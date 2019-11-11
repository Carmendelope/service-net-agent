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

package defaults

// Agent defaults

import (
	"os"
)

const (
	AgentName        = "service-net-agent"
	AgentDescription = "Nalej Service Net Agent"

	AgentHeartbeatInterval = 30
	AgentCommTimeout       = 15
	AgentShutdownTimeout   = 60
	AgentOpQueueLen        = 32
	AgentOpTimeout         = 15 // Any individual operation can take at most this long

	// Used to generate a unique but safe agent id
	ApplicationID = "allyourbasearebelongtonalej"

	ConfigFile string = "etc" + string(os.PathSeparator) + "agent.yaml"
	LogFile    string = "log" + string(os.PathSeparator) + "agent.log"
	BinDir     string = "bin"
)
