/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package plugin

// Plugin infrastructure

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/spf13/viper"
)

// Type for indexing plugins by name
type PluginName string
func (n PluginName) String() string {
	return string(n)
}

// Type for function to create a new plugin instance
type NewPluginFunc func(config *viper.Viper) (Plugin, derrors.Error)

// Type for indexing commands by name
type CommandName string
func (n CommandName) String() string {
	return string(n)
}

// Type for functions that implement execution of plugin commands
type CommandFunc func(ctx context.Context, params map[string]string) (string, derrors.Error)

// Type for mapping plugin commands to their implementation
type CommandMap map[CommandName]CommandFunc

// Default commands handled outside of the plugin
const (
	StartCommand CommandName = "start"
	StopCommand CommandName = "stop"
)

type Plugin interface {
	StartPlugin() (derrors.Error)
	StopPlugin()

	// Retrieve the function for a specific command to be executed in the
	// caller. Alternative would be an ExecuteCommand() function on
	// the plugin, but that would duplicate implementation for each plugin
	// that we can now implement in the registry.
	GetCommand(cmd CommandName) CommandFunc
}
