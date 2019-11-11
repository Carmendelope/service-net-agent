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

package commands

import (
	"time"

	"github.com/nalej/service-net-agent/internal/app/run"
	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"
	"github.com/nalej/service-net-agent/pkg/svcmgr"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var service = &run.Service{
	Config: rootConfig,
}

var runAsService bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Service Net Agent",
	Long:  "Run Service Net Agent",
	Run: func(cmd *cobra.Command, args []string) {
		setup(cmd)
		onRun()
	},
}

func init() {
	runCmd.Flags().BoolVar(&runAsService, "service", false, "Run as system service where supported")

	runCmd.Flags().Duration("interval", time.Second*time.Duration(defaults.AgentHeartbeatInterval), "Heartbeat interval")
	rootConfig.BindPFlag("agent.interval", runCmd.Flags().Lookup("interval"))

	// No command-line options, but can be specified in config file
	rootConfig.SetDefault("agent.comm_timeout", (time.Second * time.Duration(defaults.AgentCommTimeout)).String())
	rootConfig.SetDefault("agent.shutdown_timeout", (time.Second * time.Duration(defaults.AgentShutdownTimeout)).String())
	rootConfig.SetDefault("agent.opqueue_len", defaults.AgentOpQueueLen)

	rootCmd.AddCommand(runCmd)
}

func onRun() {
	log.Info().Msg("Starting Service Net Agent")

	err := service.Validate()
	if err != nil {
		Fail(err, "invalid configuration")
	}

	service.Client, err = client.FromConfig(service.Config)
	if err != nil {
		Fail(err, "unable to create edge controller client")
	}
	defer service.Client.Close()

	manager, err := svcmgr.NewManager(defaults.AgentName, service)
	if err != nil {
		Fail(err, "unable to create service manager")
	}

	err = manager.Run(runAsService)
	if err != nil {
		Fail(err, "run failed")
	}

	log.Info().Msg("Service Net Agent stopped")
}
