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

	"github.com/spf13/viper"
)

var metricsDescriptor = plugin.PluginDescriptor{
	Name: "metrics",
	Description: "System metrics collection plugin",
	NewFunc: NewMetrics,
}

type Metrics struct {
	plugin.BasePlugin
}

func init() {
	plugin.Register(&metricsDescriptor)
}

func NewMetrics(config *viper.Viper) (plugin.Plugin, derrors.Error) {
	m := &Metrics{
	}

	return m, nil
}

func (m *Metrics) GetPluginDescriptor() (*plugin.PluginDescriptor) {
        return &metricsDescriptor
}

func (m *Metrics) Beat(context.Context) (plugin.PluginHeartbeatData, derrors.Error) {
	return &MetricsData{
		Timestamp: time.Now(),
	}, nil
}
