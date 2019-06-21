/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package core

// Core plugin functionality, to control agent from Edge Controller

import (
	"context"
	"os"

	"github.com/nalej/derrors"
	"github.com/nalej/infra-net-plugin"

	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"
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

	// Config variables needed
	config *config.Config

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

func NewCore(cfg *viper.Viper) (plugin.Plugin, derrors.Error) {
	runnerI := cfg.Get("runner")
	runner, ok := runnerI.(svcmgr.Runner)
	if !ok {
		return nil, derrors.NewInvalidArgumentError("no valid runner for core plugin")
	}

	runnerCfgI := cfg.Get("config")
	runnerCfg, ok := runnerCfgI.(*config.Config)
	if !ok {
		return nil, derrors.NewInvalidArgumentError("no valid config for core plugin")
	}

	c := &Core{
		runner: runner,
		config: runnerCfg,
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
// - Disable and remove the system service
// - Delete the configuration file
// - Stop and exit
func (c *Core) uninstall(ctx context.Context, params map[string]string) (string, derrors.Error) {
	log.Info().Msg("Received uninstall command. Stopping and disabling agent.")

	// First, make sure we don't send a heartbeat or accept any
	// operation requests anymore
	c.runner.Disable()

	// Disable and remove the system service
	installer, derr := svcmgr.NewInstaller(defaults.AgentName, c.config.Path)
	if derr != nil {
		return "", derr
	}

	derr = installer.Disable()
	if derr != nil {
		return "", derr
	}

	derr = installer.Remove()
	if derr != nil {
		return "", derr
	}

	// Delete configuration
	derr = c.config.DeleteConfigFile()
	if derr != nil {
		return "", derr
	}

	// Stop the running service
	c.runner.Stop()

	// We should be stopped now, but just in case, we just
	// exit after timeout
	<-ctx.Done()
	os.Exit(0)

	// Lol.
	return "", derrors.NewInternalError("failed to stop")
}
