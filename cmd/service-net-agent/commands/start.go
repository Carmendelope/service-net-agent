
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

package commands

import (
	"github.com/nalej/service-net-agent/internal/pkg/defaults"
	"github.com/nalej/service-net-agent/pkg/svcmgr"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:    "start",
	Short:  "Request system to start Service Net Agent service",
	Long:   "Request system to start Service Net Agent service",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		setup(cmd)
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
