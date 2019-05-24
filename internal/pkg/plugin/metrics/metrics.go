/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics

// Metrics plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/cpu"
	_ "github.com/influxdata/telegraf/plugins/inputs/net"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var metricsDescriptor = plugin.PluginDescriptor{
	Name: "metrics",
	Description: "System metrics collection plugin",
	NewFunc: NewMetrics,
}

type Metrics struct {
	plugin.BasePlugin

	inputs []*RunningInput
}

func init() {
	plugin.Register(&metricsDescriptor)
}

type RunningInput struct {
	InputName string
	Input telegraf.Input
}

func NewRunningInput(name string, input telegraf.Input) *RunningInput {
	rp := &RunningInput{
		InputName: name,
		Input: input,
	}

	return rp
}

func (i *RunningInput) Name() string {
	return i.InputName
}

func (i *RunningInput) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

func (i *RunningInput) Gather(acc telegraf.Accumulator) error {
	// TODO:
	// - capture internal timing
	// - run in context?
	return i.Input.Gather(acc)
}

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
We will implement filtering if and when needed for our purposes - there is
no need to re-implement the same structure as Telegraf. For now, we just pass
the original metric.
*/

/*
plugin config models.InputConfig
- name
- interval
- name_prefix
- name_suffix
- name_override
- tags
- filter
 - namepass
 - namedrop
 - fieldpass
 - fielddrop
 - tagpass
 - tagdrop
 - tagexclude
 - taginclude
*/

func CreateTomlTable(config map[string]interface{}) (*ast.Table, derrors.Error) {
	fields := make(map[string]interface{}, len(config))

	for k, v := range(config) {
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

func NewMetrics(config *viper.Viper) (plugin.Plugin, derrors.Error) {
	inputsList := []*RunningInput{}

	name := "cpu"
	configMap := map[string]interface{}{
		"percpu": true,
		"totalcpu": false,
		"collect_cpu_time": true,
		"report_active": false,
		// include-exclude filter
	}
	configTable, derr := CreateTomlTable(configMap)
	if derr != nil {
		return nil, derr
	}

	inputCreator, found := inputs.Inputs[name]
	if !found {
		return nil, derrors.NewUnavailableError("input not available").WithParams(name)
	}

	input := inputCreator()
	err := toml.UnmarshalTable(configTable, input)
	if err != nil {
		return nil, derrors.NewInternalError("failed parsing config", err)
	}
	log.Debug().Interface("input", input).Msg("input configured")
	inputsList = append(inputsList, NewRunningInput(name, input))
	m := &Metrics{
		inputs: inputsList,
	}

	return m, nil
}

func (m *Metrics) GetPluginDescriptor() (*plugin.PluginDescriptor) {
        return &metricsDescriptor
}

func (m *Metrics) Beat(context.Context) (plugin.PluginHeartbeatData, derrors.Error) {
	metricChan := make(chan telegraf.Metric, 100)

	// We can collect and process metrics in parallel
	go func() {
		for _, input := range(m.inputs) {
			acc := agent.NewAccumulator(input, metricChan)
			err := input.Gather(acc)
			if err != nil {
				close(metricChan)
				log.Warn().Err(err).Str("input", input.Name()).Msg("error gathering input")
				return
			}
		}

		// Now we can range over it
		close(metricChan)
	}()

	// TODO context
	for metric := range(metricChan) {
		log.Debug().Str("name", metric.Name()).Interface("tags", metric.Tags()).Fields(metric.Fields()).Msg("metric")
		metric.Accept()

	}

	// CPU usage percentage since last call (0 interval),
	// total of all CPUs (false)

	// NOTE: close channel we're reading from to interrupt (inputC)

	/*
	times, err := cpu.Percent(0, false)
	if err != nil {

	}

	log.Info().Interface("cpu", times).Msg("cpu")
	*/
	return &MetricsData{
		Timestamp: time.Now(),
	}, nil
}
