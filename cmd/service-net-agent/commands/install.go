// +build linux

/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// For Windows, we have a separate installer

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
		onInstall()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func onInstall() {
	log.Info().Msg("Starting installation")

	err := installer.Validate()
	if err != nil {
		Fail(err, "invalid configuration")
	}

	err = installer.Run()
	if err != nil {
		Fail(err, "installation failed")
	}

	log.Info().Msg("Installation successfull")
}
