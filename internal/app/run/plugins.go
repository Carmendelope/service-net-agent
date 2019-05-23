/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Available plugins

import (
	_ "github.com/nalej/service-net-agent/internal/pkg/plugin/metrics"
	_ "github.com/nalej/service-net-agent/internal/pkg/plugin/ping"

	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/rs/zerolog/log"
)

func printRegisteredPlugins() {
	for name, entry := range(plugin.ListPlugins()) {
		log.Info().Str("name", name.String()).Str("description", entry.Description).Bool("running", entry.Running).Msg("plugin loaded")
	}
}
