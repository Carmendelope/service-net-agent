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

package metrics

// Metrics inputs - wrapping Telegraf input plugins

import (
	"fmt"

	"github.com/nalej/derrors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"

	_ "github.com/influxdata/telegraf/plugins/inputs/cpu"
	_ "github.com/influxdata/telegraf/plugins/inputs/disk"
	_ "github.com/influxdata/telegraf/plugins/inputs/diskio"
	_ "github.com/influxdata/telegraf/plugins/inputs/mem"
	_ "github.com/influxdata/telegraf/plugins/inputs/net"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

/*
A plugin gets instantiated by calling its creation function, which is in the
inputs.Inputs map, by plugin name. In Telegraf, the plugin-specific
configuration gets set directly into the plugin instance by parsing relevant
TOML fields; we do this by creating the intermediate parsed structure and
handing that off to the TOML Unmarshaller.

Metrics are gathered by calling the Gather(acc) method on the plugin instance,
which takes an Accumulator. The standard Accumulator (agent.Accumulator)
collects and processes a metric from a plugin and outputs it on a channel.

To create an agent.Accumulator, we need to wrap the plugin in a
models.RunningInput. We have to implement that here, because the whole models
package is internal to Telegraf.

A RunningInput only has two function calls: Name() returns the name of the
plugin, and MakeMetric() filters a metric gathered from the plugin Gather()
call before passing it on the Accumulator channel. This filter is based
on the plugin configuration. See:
- https://github.com/influxdata/telegraf/blob/master/internal/models/running_input.go
- https://github.com/influxdata/telegraf/blob/master/internal/models/makemetric.go
- https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md
  (sections on Input Plugin and Metric Filtering)
We have implemented filtering just as a list of metrics to include, that's all we
need and it makes it explicit what we actually collect.
*/

type InputConfigMap map[string]interface{}
type InputConfig struct {
	Name   string
	Config InputConfigMap
	Fields []string
}

type RunningInput struct {
	Config *InputConfig
	Input  telegraf.Input

	filter filter.Filter
}

func NewRunningInput(config *InputConfig) (*RunningInput, derrors.Error) {
	// Get plugin from Telegraf registry
	inputCreator, found := inputs.Inputs[config.Name]
	if !found {
		return nil, derrors.NewUnavailableError("input not available").WithParams(config.Name)
	}
	input := inputCreator()

	// Set config
	configTable, derr := CreateTomlTable(config.Config)
	if derr != nil {
		return nil, derr
	}
	err := toml.UnmarshalTable(configTable, input)
	if err != nil {
		return nil, derrors.NewInternalError("failed parsing config", err)
	}

	f, err := filter.Compile(config.Fields)
	if err != nil {
		return nil, derrors.NewInternalError("error creating fields filter").WithParams(config.Name)
	}

	r := &RunningInput{
		Config: config,
		Input:  input,
		filter: f,
	}

	return r, nil
}

func (i *RunningInput) Name() string {
	return i.Config.Name
}

func (i *RunningInput) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	// No filter means all fields
	if i.filter == nil {
		return metric
	}

	// Include requested fields
	filterKeys := []string{}
	for _, field := range metric.FieldList() {
		if !i.filter.Match(field.Key) {
			filterKeys = append(filterKeys, field.Key)
		}
	}
	for _, key := range filterKeys {
		metric.RemoveField(key)
	}

	// Drop empty metrics
	if len(metric.FieldList()) == 0 {
		metric.Drop()
		return nil
	}

	return metric
}

func (i *RunningInput) Gather(metricChan chan telegraf.Metric) error {
	acc := agent.NewAccumulator(i, metricChan)
	return i.Input.Gather(acc)
}

func CreateTomlTable(config map[string]interface{}) (*ast.Table, derrors.Error) {
	fields := make(map[string]interface{}, len(config))

	for k, v := range config {
		var astVal ast.Value
		strVal := fmt.Sprint(v)

		switch t := v.(type) {
		case bool:
			astVal = &ast.Boolean{Value: strVal}
		default:
			return nil, derrors.NewInvalidArgumentError("unknown type in config").WithParams(t)
		}

		fields[k] = &ast.KeyValue{Value: astVal}
	}

	table := &ast.Table{
		Fields: fields,
	}

	return table, nil
}
