/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Agent configuration file management

package config

import (
	"github.com/nalej/derrors"
)

type Config struct {
	ConfigFile string
}

func (c *Config) Read() (derrors.Error) {
	return nil
}
