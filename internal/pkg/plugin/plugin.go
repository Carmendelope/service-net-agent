/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package plugin

// Plugin infrastructure

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/spf13/viper"
)

// Type used to describe plugins and how to instantiate them
type PluginDescriptor struct {
	Name PluginName
	Description string

	// Function used to instantiate a plugin
	NewFunc NewPluginFunc

	// All commands supported by the plugin
	Commands CommandMap
}

func (d *PluginDescriptor) AddCommand(command CommandDescriptor) {
	if d.Commands == nil {
		d.Commands = CommandMap{}
	}

	d.Commands[command.Name] = command
}

// Type for indexing plugins by name
type PluginName string
func (n PluginName) String() string {
	return string(n)
}

// Type for function to create a new plugin instance
type NewPluginFunc func(config *viper.Viper) (Plugin, derrors.Error)

// Type used to describe a plugin command
type CommandDescriptor struct {
	Name CommandName
	Description string

	Params ParamMap
}

func (d *CommandDescriptor) AddParam(param ParamDescriptor) {
	if d.Params == nil {
		d.Params = ParamMap{}
	}

	d.Params[param.Name] = param
}

// Type for indexing commands by name
type CommandName string
func (n CommandName) String() string {
	return string(n)
}

// Type for functions that implement execution of plugin commands
type CommandFunc func(ctx context.Context, params map[string]string) (string, derrors.Error)

// Type for mapping plugin commands to their implementation
type CommandMap map[CommandName]CommandDescriptor
type CommandFuncMap map[CommandName]CommandFunc

// Type for indexing parameters by name
type ParamName string
func (n ParamName) String() string {
	return string(n)
}

// Type to describe plugin command parameters
type ParamDescriptor struct {
	Name ParamName
	Description string

	Required bool
	Default string
}

type ParamMap map[ParamName]ParamDescriptor

// Default commands handled outside of the plugin
const (
	StartCommand CommandName = "start"
	StopCommand CommandName = "stop"
)

type Plugin interface {
	StartPlugin() (derrors.Error)
	StopPlugin()

	// Retrieve the plugin descriptor
	GetPluginDescriptor() *PluginDescriptor

	// Callback for plugin to add data to heartbeat. Returns nil if
	// nothing needs to be added
	Beat(ctx context.Context) (PluginHeartbeatData, derrors.Error)

	// Retrieve the function for a specific command to be executed in the
	// caller. Alternative would be an ExecuteCommand() function on
	// the plugin, but that would duplicate implementation for each plugin
	// that we can now implement in the registry.
	GetCommandFunc(cmd CommandName) CommandFunc
}

type PluginHeartbeatData interface {
	ToGRPC() *grpc_edge_controller_go.PluginData
}

type PluginHeartbeatDataList []PluginHeartbeatData
func (d PluginHeartbeatDataList) ToGRPC() []*grpc_edge_controller_go.PluginData {
	out := make ([]*grpc_edge_controller_go.PluginData, 0, len(d))
	for _, data := range(d) {
		out = append(out, data.ToGRPC())
	}

	return out
}
