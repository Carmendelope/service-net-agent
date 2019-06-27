/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics

import (
	"time"

	"github.com/nalej/grpc-edge-controller-go"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("metricsdata", func() {
	ginkgo.It("should not create metrics with no fields", func() {
		gomega.Expect(NewMetric(testName, nil, nil)).To(gomega.BeNil())
	})
	ginkgo.It("should create metrics with tags", func() {
		gomega.Expect(NewMetric(testName, testTags, testFields)).To(gomega.Equal(testMetric))
	})
	ginkgo.It("should create a gRPC message from a metric", func() {
		ts := time.Now()
		data := &MetricsData{
			Timestamp: ts,
			Metrics: []*Metric{testMetric, testMetric},
		}

		expected := &grpc_edge_controller_go.PluginData{
			Plugin: 0,
			Data: &grpc_edge_controller_go.PluginData_MetricsData{
				MetricsData: &grpc_edge_controller_go.MetricsPluginData{
					Timestamp: ts.UTC().Unix(),
					Metrics: []*grpc_edge_controller_go.MetricsPluginData_Metric{
						&grpc_edge_controller_go.MetricsPluginData_Metric{
							Name: testName,
							Tags: testTags,
							Fields: testMetric.Fields,
						},
						&grpc_edge_controller_go.MetricsPluginData_Metric{
							Name: testName,
							Tags: testTags,
							Fields: testMetric.Fields,
						},
					},
				},
			},
		}

		gomega.Expect(data.ToGRPC()).To(gomega.Equal(expected))
	})

	ginkgo.Context("when converting values", func() {
		ginkgo.It("should create metrics from uint64 fields", func() {
			fields := map[string]interface{}{
				"field1": uint64(0),
				"field2": uint64(12345),
				"field3": uint64(1<<64 - 1), // max uint64
			}
			expected := map[string]uint64{
				"field1": 0,
				"field2": 12345,
				"field3": 18446744073709551615,
			}
			metric, _:= NewMetric(testName, nil, fields)
			gomega.Expect(metric.Fields).To(gomega.Equal(expected))
		})
		ginkgo.It("should create metrics from float64 fields", func() {
			fields := map[string]interface{}{
				"field1": float64(-10),
				"field2": float64(0.3),
				"field3": float64(1.8),
				"field4": float64(1 << 63 - 1), // max int64
				"field5": float64(18446744073709551616), // max uint64 + 1
			}
			expected := map[string]uint64{
				"field2": 0,
				"field3": 2,
				"field4": 9223372036854775808,
			}
			metric, _:= NewMetric(testName, nil, fields)
			gomega.Expect(metric.Fields).To(gomega.Equal(expected))
		})
		ginkgo.It("should fail on fields of unknown type", func() {

		})
	})
})
