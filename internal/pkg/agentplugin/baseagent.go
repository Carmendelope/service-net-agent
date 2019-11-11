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

package agentplugin

// Base plugin that can be embedded in other plugins to adhere to the interface

import (
	"context"

	"github.com/nalej/derrors"
	"github.com/nalej/infra-net-plugin"
)

type BaseAgentPlugin struct {
	plugin.BasePlugin
}

// A plugin can embed BaseAgentPlugin and only has to define GetPluginDescriptor
// func (b *BasePlugin) GetPluginDescriptor() *PluginDescriptor

// No heartbeat data
func (b *BaseAgentPlugin) Beat(context.Context) (PluginHeartbeatData, derrors.Error) {
	return nil, nil
}
