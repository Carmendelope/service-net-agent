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

// Metrics return data

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/rs/zerolog/log"
)

type MetricsData struct {
	Timestamp time.Time
	Metrics   []*Metric
}

type Metric struct {
	Name   string
	Tags   map[string]string
	Fields map[string]uint64
}

func NewMetric(name string, tags map[string]string, fields map[string]interface{}) (*Metric, derrors.Error) {
	// Skip empty metrics
	if len(fields) == 0 {
		return nil, nil
	}

	metric := &Metric{
		Name:   name,
		Tags:   tags,
		Fields: make(map[string]uint64, len(fields)),
	}

	for key, val := range fields {
		var intVal uint64
		switch v := val.(type) {
		// Supporting plugins with different value types requires
		// additional types here
		case uint64:
			intVal = v
		case float64:
			var derr derrors.Error
			intVal, derr = floatToUint(v)
			if derr != nil {
				log.Warn().Err(derr).Float64("val", v).Msg("field value out of bounds")
				continue
			}
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
	for _, m := range d.Metrics {
		grpcMetric := &grpc_edge_controller_go.MetricsPluginData_Metric{
			Name:   m.Name,
			Tags:   m.Tags,
			Fields: m.Fields,
		}
		grpcMetrics = append(grpcMetrics, grpcMetric)
	}

	data := &grpc_edge_controller_go.PluginData{
		Plugin: grpc_edge_controller_go.Plugin(grpcPlugin),
		Data: &grpc_edge_controller_go.PluginData_MetricsData{
			MetricsData: &grpc_edge_controller_go.MetricsPluginData{
				Timestamp: d.Timestamp.UTC().Unix(),
				Metrics:   grpcMetrics,
			},
		},
	}

	return data
}

func floatToUint(f float64) (uint64, derrors.Error) {
	if f < 0 || f > math.MaxInt64 { // After MaxInt64 conversions break it seems
		return 0, derrors.NewInvalidArgumentError("out-of-bounds conversion from float64 to uint64").WithParams(f)
	}

	i := uint64(math.Round(f))
	return i, nil
}
