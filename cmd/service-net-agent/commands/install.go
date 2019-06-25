// +build linux windows

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
		setup(cmd)
		onInstall(install.InstallCommand)
	},
}

var uninstallCmd = &cobra.Command{
	Use: "uninstall",
	Short: "Install Service Net Agent",
	Long: "Install Service Net Agent",
	Run: func(cmd *cobra.Command, args []string) {
		setup(cmd)
		onInstall(install.UninstallCommand)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
}

func onInstall(cmd install.InstallCommandType) {
	log.Info().Msg("Starting " + cmd.String())

	err := installer.Validate()
	if err != nil {
		Fail(err, "invalid configuration")
	}

	err = installer.Run(cmd)
	if err != nil {
		Fail(err, cmd.String() + " failed")
	}

	log.Info().Msg(cmd.String() + " successfull")
}
