/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
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
	Use: "join",
	Short: "Join Service Net Agent",
	Long: "Join Service Net Agent to Nalej Edge through Edge Controller",
	Run: func(cmd *cobra.Command, args []string) {
		Setup(cmd)
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
