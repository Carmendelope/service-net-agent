/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
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
