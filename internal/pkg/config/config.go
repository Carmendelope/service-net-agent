/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Agent configuration file management

package config

import (
	"os"

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

func (c *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	for _, key := range(c.AllKeys()) {
		log.Info().Interface(key, c.Get(key)).Msg("configuration value")
	}
}
