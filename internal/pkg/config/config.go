/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Agent configuration file management

package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/version"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Path string
	ConfigFile string

	*viper.Viper
}

func NewConfig() (*Config) {
	c := &Config{
		Viper: viper.New(),
	}

	// We store a token, make it only readable for user
	// See comment in Write()
	// c.Viper.SetConfigPermissions(0600)

	return c
}

func (c *Config) Read() (derrors.Error) {
	log.Info().Str("file", c.ConfigFile).Msg("reading configuration file")

	// Pass filename to Viper
	c.SetConfigFile(c.ConfigFile)

	// Check if file exists. Ok if not, just don't try to read.
	if _, err := os.Stat(c.ConfigFile); os.IsNotExist(err) {
		return nil
	}

	// Read
	err := c.ReadInConfig()
	if err != nil {
		return derrors.NewInvalidArgumentError("failed reading configuration file", err)
	}

	return nil
}

func (c *Config) Write() derrors.Error {
	confDir := filepath.Dir(c.ConfigFile)
	err := os.MkdirAll(confDir, 0755)
	if err != nil {
		return derrors.NewPermissionDeniedError("failed creating config dir", err).WithParams(confDir)
	}

	// Unstable version of viper allows to set filemode, we need
	// to do it after writing. This does introduce a slight vulnerability
	// as there is a small window during which the file can be read.
	// This will be fixed with the next version of viper.
	err = c.WriteConfig()
	if err != nil {
		return derrors.NewInternalError("failed writing config file", err).WithParams(c.ConfigFile)
	}

	err = os.Chmod(c.ConfigFile, 0600)
	if err != nil {
		return derrors.NewInternalError("failed setting config file permissions", err).WithParams(c.ConfigFile)
	}

	return nil
}

func (c *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	for _, key := range(c.AllKeys()) {
		val := c.Get(key)

		// Don't print secrets
		if strings.Contains(key, "token") {
			val = interface{}(strings.Repeat("*", len(val.(string))))
		}

		log.Info().Interface(key, val).Msg("configuration value")
	}
}
