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
	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/defaults"

	"github.com/nalej/service-net-agent/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	debugLevel     bool
	consoleLogging bool
	logToFile      bool
	logFile        string
	logOut         io.Writer
	rootConfig     = config.NewConfig()
)

var rootCmd = &cobra.Command{
	Use:     defaults.AgentName,
	Short:   defaults.AgentDescription,
	Long:    `Nalej Service Net Agent joins resources to the Nalej Edge`,
	Version: "unknown-version",

	Run: func(cmd *cobra.Command, args []string) {
		setup(cmd)
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

	rootCmd.PersistentFlags().Bool("tls", true, "Use TLS to connect to Edge Controller")
	rootConfig.BindPFlag("controller.tls", rootCmd.PersistentFlags().Lookup("tls"))

	rootCmd.PersistentFlags().Bool("insecure", false, "Don't check Edge Controller certificate")
	rootConfig.BindPFlag("controller.insecure", rootCmd.PersistentFlags().Lookup("insecure"))

	rootCmd.PersistentFlags().String("cert", "", "File with certificate to use to connect to Edge Controller")
	rootConfig.BindPFlag("controller.cert", rootCmd.PersistentFlags().Lookup("cert"))

	rootCmd.PersistentFlags().BoolVar(&logToFile, "log", false, "Output to log file")
	rootCmd.PersistentFlags().StringVar(&logFile, "logfile", filepath.Join(rootConfig.Path, defaults.LogFile), "File to write log to if --log is set")
	rootCmd.MarkPersistentFlagFilename("logfile")

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
func setup(cmd *cobra.Command) {
	// Logging
	// If log file is not specifically set, we should look in root path
	if !cmd.Flag("logfile").Changed {
		logFile = filepath.Join(rootConfig.Path, defaults.LogFile)
	}
	err := setupLogging(logToFile, logFile)
	if err != nil {
		Fail(err, "failed setting up logging")
	}

	// Set logfile in config - used to properly set command line parameters
	// for Windows (where we enable logging). We set this even if currently
	// logging to file is not enabled, because we want to always set it
	// when starting as a service.
	// Ideally we would use the Windows event logger, but that requires
	// wrapping ZeroLog, which is a whole different ballgame.
	rootConfig.LogFile = logFile

	// If config file is not specifically set, we should look in root path
	if !cmd.Flag("config").Changed {
		rootConfig.ConfigFile = filepath.Join(rootConfig.Path, defaults.ConfigFile)
	}

	// Handle configuration
	err = rootConfig.Read()
	if err != nil {
		Fail(err, "failed to read configuration file")
	}
}

// setupLogging sets the debug level and console logging if required.
func setupLogging(toFile bool, file string) derrors.Error {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugLevel {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if toFile {
		err := os.MkdirAll(filepath.Dir(file), 0755)
		if err != nil {
			return derrors.NewPermissionDeniedError("unable to create log file directory", err).WithParams(file)
		}
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return derrors.NewPermissionDeniedError("unable to open log file for writing", err).WithParams(file)
		}
		logOut = f
	} else {
		logOut = os.Stderr
	}

	if consoleLogging {
		logOut = zerolog.ConsoleWriter{Out: logOut}
	}

	log.Logger = log.Output(logOut)

	return nil
}

func Fail(err derrors.Error, msg string) {
	var logger zerolog.Logger = log.Logger

	// Make sure we print to stderr
	if logToFile && logOut != nil {
		var console io.Writer = os.Stderr
		if consoleLogging {
			console = zerolog.ConsoleWriter{Out: console}
		}
		out := io.MultiWriter(logOut, console)
		logger = log.Output(out)
	}

	logger.Debug().Str("err", err.DebugReport()).Msg("debug report")
	logger.Fatal().Err(err).Msg(msg)
}
