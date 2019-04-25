// +build linux

/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/service-net-agent/internal/pkg/defaults"
	"github.com/nalej/service-net-agent/pkg/svcmgr"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use: "start",
	Short: "Request system to start Service Net Agent service",
	Long: "Request system to start Service Net Agent service",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		Setup(cmd)
		onStart()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func onStart() {
	log.Info().Msg("Requesting system to start Service Net Agent")

	err := svcmgr.Start(defaults.AgentName)
	if err != nil {
		Fail(err, "unable to start service net agent")
	}

	// Result is printed by Start function.
}
