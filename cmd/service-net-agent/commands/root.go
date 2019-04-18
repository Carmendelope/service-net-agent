/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"

	"github.com/nalej/service-net-agent/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"fmt"
	"os"
)

var debugLevel bool
var consoleLogging bool
var configuration = &config.Config{}

var rootCmd = &cobra.Command{
	Use:   "service-net-agent",
	Short: "Nalej Service Net Agent",
	Long:  `Nalej Service Net Agent joins resources to the Nalej Edge`,
	Version: "unknown-version",

	Run: func(cmd *cobra.Command, args []string) {
		Setup()
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configuration.ConfigFile, "config", defaults.ConfigFile, "Configuration file")
	rootCmd.PersistentFlags().BoolVar(&debugLevel, "debug", false, "Set debug level")
	rootCmd.PersistentFlags().BoolVar(&consoleLogging, "consoleLogging", false, "Pretty print logging")
}

func Execute() {
	rootCmd.SetVersionTemplate(version.GetVersionInfo())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Initialize the agent
func Setup() {
	// Logging
	SetupLogging()

	// Config file
	err := configuration.Read()
	if err != nil {
		log.Fatal().Err(err).Str("config", configuration.ConfigFile).Msg("failed to read configuration file")
	}
}

// SetupLogging sets the debug level and console logging if required.
func SetupLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugLevel {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if consoleLogging {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
