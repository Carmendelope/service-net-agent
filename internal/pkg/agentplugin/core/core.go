/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package core

// Core plugin functionality, to control agent from Edge Controller

import (
	"context"

	"github.com/nalej/derrors"
	"github.com/nalej/infra-net-plugin"

	"github.com/nalej/service-net-agent/pkg/svcmgr"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var coreDescriptor = plugin.PluginDescriptor{
	Name: "core",
	Description: "Core agent control",
	NewFunc: NewCore,
}

type Core struct {
	// Note: this is not an agent plugin as it doesn't have a heartbeat
	// callback function
	plugin.BasePlugin

	// To control the running service
	runner svcmgr.Runner

	commandMap plugin.CommandFuncMap
}

func init() {
	uninstallCmd := plugin.CommandDescriptor{
		Name: "uninstall",
		Description: "stop and disable agent",
	}
	coreDescriptor.AddCommand(uninstallCmd)

	plugin.Register(&coreDescriptor)
}

func NewCore(config *viper.Viper) (plugin.Plugin, derrors.Error) {
	runnerI := config.Get("runner")
	runner, ok := runnerI.(svcmgr.Runner)
	if !ok {
		return nil, derrors.NewInvalidArgumentError("no valid runner for core plugin")
	}

	c := &Core{
		runner: runner,
	}

	c.commandMap = plugin.CommandFuncMap{
		"uninstall": c.uninstall,
	}


	return c, nil
}

func (c *Core) GetPluginDescriptor() (*plugin.PluginDescriptor) {
	return &coreDescriptor
}

func (c *Core) GetCommandFunc(cmd plugin.CommandName) plugin.CommandFunc {
	return c.commandMap[cmd]
}

// Uninstall command will:
// - Disable sending of heartbeat and receiving/executing of commands
// - Disable the system service
// - Delete the configuration file
// - Stop the system service
// - (And if the above doesn't exit the agent because it's not running as a
//    system service) Exit
func (c *Core) uninstall(ctx context.Context, params map[string]string) (string, derrors.Error) {
	log.Info().Msg("Received uninstall command. Stopping and disabling agent.")

	return "", nil
}
