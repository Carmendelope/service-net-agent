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
		UseTLS:   config.GetBool("controller.tls"),
		CACert:   config.GetString("controller.cert"),
		Insecure: config.GetBool("controller.insecure"),
		Timeout:  config.GetDuration("agent.comm_timeout"),
		Token:    t,
	}

	return NewAgentClient(config.GetString("controller.address"), opts)
}
