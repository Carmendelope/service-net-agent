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

package run

// Available plugins

import (
	_ "github.com/nalej/infra-net-plugin/ping"
	_ "github.com/nalej/service-net-agent/internal/pkg/agentplugin/core"
	_ "github.com/nalej/service-net-agent/internal/pkg/agentplugin/metrics"

	"github.com/nalej/infra-net-plugin"

	"github.com/rs/zerolog/log"
)

func printRegisteredPlugins() {
	for name, entry := range plugin.ListPlugins() {
		log.Info().Str("name", name.String()).Str("description", entry.Description).Bool("running", entry.Running).Msg("plugin loaded")
	}
}
