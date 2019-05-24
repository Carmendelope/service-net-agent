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
	"github.com/influxdata/telegraf/filter"
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
	Config *InputConfig
	Input telegraf.Input

	filter filter.Filter
}

func NewRunningInput(input telegraf.Input, config *InputConfig) (*RunningInput, derrors.Error) {
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
		Input: input,
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
	for _, field := range(metric.FieldList()) {
		if !i.filter.Match(field.Key) {
			filterKeys = append(filterKeys, field.Key)
		}
	}
	for _, key := range(filterKeys) {
		metric.RemoveField(key)
	}

	// Drop empty metrics
	if len(metric.FieldList()) == 0 {
		metric.Drop()
		return nil
	}

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
We have implemented filtering just as a list of metrics to include, that's all we
need and it makes it explicit what we actually collect.
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

type InputConfig struct {
	Name string
	Config InputConfigMap
	Fields []string
}

type InputConfigMap map[string]interface{}

var Inputs = []InputConfig{
	InputConfig{
		Name: "cpu",
		Config: InputConfigMap{
			"percpu": true,
			"totalcpu": false,
			"collect_cpu_time": true,
			"report_active": false,
		},
		Fields: []string{
			"time_*",
		},
	},
	InputConfig{
		Name: "net",
	},
}


func NewMetrics(config *viper.Viper) (plugin.Plugin, derrors.Error) {
	inputsList := make([]*RunningInput, 0, len(Inputs))

	for _, config := range(Inputs) {
		// Get plugin from Telegraf registry
		inputCreator, found := inputs.Inputs[config.Name]
		if !found {
			return nil, derrors.NewUnavailableError("input not available").WithParams(config.Name)
		}
		input := inputCreator()

		running, derr := NewRunningInput(input, &config)
		if derr != nil {
			return nil, derr
		}

		log.Debug().Interface("input", input).Msg("input configured")
		inputsList = append(inputsList, running)
	}

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
