/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"

	"github.com/nalej/service-net-agent/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"fmt"
	"os"
	"path/filepath"
)

var debugLevel bool
var consoleLogging bool
var rootConfig = config.NewConfig()

var rootCmd = &cobra.Command{
	Use:   defaults.AgentName,
	Short: defaults.AgentDescription,
	Long:  `Nalej Service Net Agent joins resources to the Nalej Edge`,
	Version: "unknown-version",

	Run: func(cmd *cobra.Command, args []string) {
		Setup(cmd)
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootConfig.Path, "path", defaults.Path, "Agent root path")
	rootCmd.MarkPersistentFlagFilename("path")

	rootCmd.PersistentFlags().StringVar(&rootConfig.ConfigFile, "config", filepath.Join(rootConfig.Path, defaults.ConfigFile), "Configuration file")
	rootCmd.MarkPersistentFlagFilename("config")

	rootCmd.PersistentFlags().String("address", "", "Edge Controller address")
	rootConfig.BindPFlag("controller.address", rootCmd.PersistentFlags().Lookup("address"))

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
func Setup(cmd *cobra.Command) {
	// Logging
	SetupLogging()

	// If config file is not specifically set, we should look in root path
	if !cmd.Flag("config").Changed {
		rootConfig.ConfigFile = filepath.Join(rootConfig.Path, defaults.ConfigFile)
	}

	// Handle configuration
	err := rootConfig.Read()
	if err != nil {
		Fail(err, "failed to read configuration file")
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

func Fail(err derrors.Error, msg string) {
	log.Debug().Str("err", err.DebugReport()).Msg("debug report")
	log.Fatal().Err(err).Msg(msg)
}
