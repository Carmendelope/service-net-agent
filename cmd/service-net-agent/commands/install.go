
// +build linux windows

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
	Use:   "install",
	Short: "Install Service Net Agent",
	Long:  "Install Service Net Agent",
	Run: func(cmd *cobra.Command, args []string) {
		setup(cmd)
		onInstall(install.InstallCommand)
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Install Service Net Agent",
	Long:  "Install Service Net Agent",
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
		Fail(err, cmd.String()+" failed")
	}

	log.Info().Msg(cmd.String() + " successfull")
}
