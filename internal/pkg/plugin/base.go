/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package plugin

// Base plugin that can be embedded in other plugins to adhere to the interface

import (
	"context"

	"github.com/nalej/derrors"
)

type BasePlugin struct { }

// A plugin can embed BasePlugin and only has to define GetPluginDescriptor
// func (b *BasePlugin) GetPluginDescriptor() *PluginDescriptor

// Don't do anything specific on start
func (b *BasePlugin) StartPlugin() (derrors.Error) {
	return nil
}

// Don't do anything specific on stop
func (b *BasePlugin) StopPlugin() {
}

// No commands
func (b *BasePlugin) GetCommandFunc(CommandName) CommandFunc {
	return nil
}

// No heartbeat data
func (b *BasePlugin) Beat(context.Context) (PluginHeartbeatData, derrors.Error) {
	return nil, nil
}
