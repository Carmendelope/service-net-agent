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

package core

// Core plugin functionality, to control agent from Edge Controller

import (
	"context"
	"os"
	"time"

	"github.com/nalej/derrors"
	"github.com/nalej/infra-net-plugin"

	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"
	"github.com/nalej/service-net-agent/pkg/svcmgr"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var coreDescriptor = plugin.PluginDescriptor{
	Name:        "core",
	Description: "Core agent control",
	NewFunc:     NewCore,
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
		Name:        "uninstall",
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

func (c *Core) GetPluginDescriptor() *plugin.PluginDescriptor {
	return &coreDescriptor
}

func (c *Core) GetCommandFunc(cmd plugin.CommandName) plugin.CommandFunc {
	return c.commandMap[cmd]
}

// Uninstall command will:
// - Run actual uninstallation in a Go routine
// - Return such that the Edge Controller gets confirmation of command
// The actual uninstallation will:
// - Disable sending of heartbeat and receiving/executing of commands
// - Disable and remove the system service
// - Delete the configuration file
// - Stop and exit
func (c *Core) uninstall(ctx context.Context, params map[string]string) (string, derrors.Error) {
	log.Info().Msg("Received uninstall command. Stopping and disabling agent.")

	go c.doUninstall(ctx)
	return "Uninstall in progress", nil
}

func (c *Core) doUninstall(ctx context.Context) {
	log.Debug().Msg("executing uninstall")

	// First, make sure we don't send a heartbeat or accept any
	// operation requests anymore
	c.runner.Disable()

	// Disable and remove the system service
	installer, derr := svcmgr.NewInstaller(defaults.AgentName, c.config.Path)
	if derr != nil {
		log.Debug().Str("trace", derr.DebugReport()).Msg("debug report")
		log.Error().Err(derr).Msg("failed to create installer")
		return
	}

	derr = installer.Disable()
	if derr != nil {
		log.Debug().Str("trace", derr.DebugReport()).Msg("debug report")
		log.Warn().Err(derr).Msg("continuing")
	}

	derr = installer.Remove()
	if derr != nil {
		log.Debug().Str("trace", derr.DebugReport()).Msg("debug report")
		log.Warn().Err(derr).Msg("continuing")
	}

	// Delete configuration
	derr = c.config.DeleteConfigFile()
	if derr != nil {
		log.Debug().Str("trace", derr.DebugReport()).Msg("debug report")
		log.Warn().Err(derr).Msg("continuing")
	}

	// Stop the running service
	c.runner.Stop()

	// We should be stopped now, but just in case, we just
	// exit after timeout. The agent most likely already stops before this
	// timeout
	time.Sleep(time.Second * defaults.AgentShutdownTimeout)
	os.Exit(0)
}
