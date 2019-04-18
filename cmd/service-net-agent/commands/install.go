/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/service-net-agent/internal/app/install"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var installer = &install.Installer{
	Config: rootConfig,
}

var installCmd = &cobra.Command{
	Use: "install",
	Short: "Install Service Net Agent",
	Long: "Install Service Net Agent",
	Run: func(cmd *cobra.Command, args []string) {
		Setup(cmd)
		Install()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func Install() {
	log.Info().Msg("Starting installation")

	err := installer.Run()
	if err != nil {
		Fail(err, "installation failed")
	}

	log.Info().Msg("Installation successfull")
}
