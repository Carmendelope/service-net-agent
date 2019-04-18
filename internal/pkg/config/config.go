/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Agent configuration file management

package config

import (
	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
)

type Config struct {
	Path string
	ConfigFile string
}

func (c *Config) Read() (derrors.Error) {
	log.Info().Str("file", c.ConfigFile).Msg("reading configuration file")
	return nil
}
