/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics

// Metrics return data

import (
	"fmt"
	"strings"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/rs/zerolog/log"
)

type MetricsData struct {
	Timestamp time.Time
	Metrics []*Metric
}

type Metric struct {
	Name string
	Tags map[string]string
	Fields map[string]uint64
}

func NewMetric(name string, tags map[string]string, fields map[string]interface{}) (*Metric, derrors.Error) {
	metric := &Metric{
		Name: name,
		Tags: tags,
		Fields: make(map[string]uint64, len(fields)),
	}

	for key, val := range(fields) {
		var intVal uint64
		switch v := val.(type) {
		// Supporting plugins with different value types requires
		// additional types here
		case uint64:
			intVal = v
		case float64:
			intVal = uint64(v)
		default:
			return nil, derrors.NewInternalError("invalid metric type").WithParams(name, fmt.Sprintf("%T", v))
		}
		metric.Fields[key] = intVal
	}

	return metric, nil
}

func (d *MetricsData) ToGRPC() *grpc_edge_controller_go.PluginData {
	ucPlugin := strings.ToUpper(metricsDescriptor.Name.String())
	grpcPlugin, ok := grpc_edge_controller_go.Plugin_value[ucPlugin]
	if !ok {
		log.Warn().Str("plugin", ucPlugin).Msg("plugin not defined in protocol")
		return nil
	}

	grpcMetrics := make([]*grpc_edge_controller_go.MetricsPluginData_Metric, 0, len(d.Metrics))
	for _, m := range(d.Metrics) {
		grpcMetric := &grpc_edge_controller_go.MetricsPluginData_Metric{
			Name: m.Name,
			Tags: m.Tags,
			Fields: m.Fields,
		}
		grpcMetrics = append(grpcMetrics, grpcMetric)
	}

	data := &grpc_edge_controller_go.PluginData{
		Plugin: grpc_edge_controller_go.Plugin(grpcPlugin),
		Data: &grpc_edge_controller_go.PluginData_MetricsData{
			MetricsData: &grpc_edge_controller_go.MetricsPluginData{
				Timestamp: d.Timestamp.UTC().Unix(),
				Metrics: grpcMetrics,
			},
		},
	}

	return data
}
