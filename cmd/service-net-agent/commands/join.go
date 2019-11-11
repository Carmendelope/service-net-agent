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
	"github.com/nalej/service-net-agent/internal/app/join"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var joiner = &join.Joiner{
	Config: rootConfig,
}

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join Service Net Agent",
	Long:  "Join Service Net Agent to Nalej Edge through Edge Controller",
	Run: func(cmd *cobra.Command, args []string) {
		setup(cmd)
		onJoin()
	},
}

func init() {
	joinCmd.Flags().StringVar(&joiner.Token, "token", "", "Join token")
	joinCmd.MarkFlagRequired("token")

	joinCmd.Flags().StringToStringVar(&joiner.Labels, "label", nil, "Asset labels")

	rootCmd.AddCommand(joinCmd)
}

func onJoin() {
	log.Info().Msg("Joining Nalej Edge")

	err := joiner.Validate()
	if err != nil {
		Fail(err, "invalid configuration")
	}

	err = joiner.Run()
	if err != nil {
		Fail(err, "join failed")
	}

	log.Info().Msg("Successfully joined Nalej Edge")
}
