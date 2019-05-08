/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package install

import (
	"github.com/nalej/service-net-agent/internal/pkg/config"
)

// Additional arguments based on config

func extraArgs(conf *config.Config) []string {
	if conf.LogFile != "" {
		return []string{
			"--log",
			"--logfile",
			conf.LogFile,
		}
	}

	return nil
}
