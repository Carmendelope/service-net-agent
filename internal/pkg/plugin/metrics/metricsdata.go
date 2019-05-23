/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics

// Metrics return data

import (
	"strings"
	"time"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/rs/zerolog/log"
)

type MetricsData struct {
	Timestamp time.Time
	Metrics map[string]int64
}

func (d *MetricsData) ToGRPC() *grpc_edge_controller_go.PluginData {
	ucPlugin := strings.ToUpper(metricsDescriptor.Name.String())
	grpcPlugin, ok := grpc_edge_controller_go.Plugin_value[ucPlugin]
	if !ok {
		log.Warn().Str("plugin", ucPlugin).Msg("plugin not defined in protocol")
		return nil
	}

	data := &grpc_edge_controller_go.PluginData{
		Plugin: grpc_edge_controller_go.Plugin(grpcPlugin),
		Data: &grpc_edge_controller_go.PluginData_MetricsData{
			MetricsData: &grpc_edge_controller_go.MetricsPluginData{
				Timestamp: d.Timestamp.UTC().Unix(),
				Metrics: d.Metrics,
			},
		},
	}

	return data
}
