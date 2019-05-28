/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics

// Metrics plugin

import (
	"context"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/influxdata/telegraf"

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

func init() {
	plugin.Register(&metricsDescriptor)
}

func NewMetrics(config *viper.Viper) (plugin.Plugin, derrors.Error) {
	inputsList := make([]*RunningInput, 0, len(Inputs))

	for _, config := range(Inputs) {
		running, derr := NewRunningInput(&config)
		if derr != nil {
			return nil, derr
		}

		log.Debug().Interface("input", running.Input).Msg("input configured")
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
			err := input.Gather(metricChan)
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

	return &MetricsData{
		// TBD
		Timestamp: time.Now(),
	}, nil
}
