/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics

// Metrics plugin

import (
	"context"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/pkg/plugin"
	"github.com/nalej/service-net-agent/internal/pkg/agentplugin"

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
	agentplugin.BaseAgentPlugin

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
		Name: "disk",
		Fields: []string{
			"used",
		},
	},
	InputConfig{
		Name: "diskio",
		Fields: []string{
			"reads", "writes", "read_bytes", "write_bytes",
		},
	},
	InputConfig{
		Name: "mem",
		Fields: []string{
			"used",
		},
	},
	InputConfig{
		Name: "net",
		Config: InputConfigMap{
			"ignore_protocol_stats": true,
		},
		Fields: []string{
			"bytes_recv", "bytes_sent", "packets_recv", "packets_sent",
		},
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

func (m *Metrics) Beat(ctx context.Context) (agentplugin.PluginHeartbeatData, derrors.Error) {
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

		// End of metrics for this beat
		close(metricChan)
	}()

	beatMetrics := make([]*Metric, 0, len(metricChan))
	for ctx.Err() == nil && metricChan != nil {
		select {
		case metric, ok := <-metricChan:
			// End of queue
			if !ok {
				metricChan = nil
				break
			}

			// Process metric
			log.Debug().Str("name", metric.Name()).Interface("tags", metric.Tags()).Fields(metric.Fields()).Msg("metric")
			beatMetric, derr := NewMetric(metric.Name(), metric.Tags(), metric.Fields())
			if derr != nil {
				return nil, derr
			}

			if beatMetric != nil {
				beatMetrics = append(beatMetrics, beatMetric)
			}
			metric.Accept()
		case <-ctx.Done():
			return nil, derrors.NewDeadlineExceededError("collecting metrics timed out", ctx.Err())
		}
	}

	return &MetricsData{
		Timestamp: time.Now(),
		Metrics: beatMetrics,
	}, nil
}
