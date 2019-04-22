/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/service-net-agent/internal/app/run"
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
	Use: "run",
	Short: "Start Service Net Agent",
	Long: "Start Service Net Agent",
	Run: func(cmd *cobra.Command, args []string) {
		Setup(cmd)
		onRun()
	},
}

func init() {
	runCmd.Flags().BoolVar(&runAsService, "service", false, "Run as system service where supported")

	runCmd.Flags().Int("interval", 30, "Heartbeat interval")
	rootConfig.BindPFlag("agent.interval", runCmd.Flags().Lookup("interval"))

	rootCmd.AddCommand(runCmd)
}

func onRun() {
	log.Info().Msg("Starting Service Net Agent")

	err := service.Validate()
	if err != nil {
		Fail(err, "invalid configuration")
	}

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
