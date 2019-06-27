/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics

import (
	"context"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("metrics", func() {
	ginkgo.It("should generate expected measurements", func() {
		p, derr := NewMetrics(nil)
		gomega.Expect(derr).To(gomega.Succeed())

		// Check if we created all inputs
		m := p.(*Metrics)
		gomega.Expect(len(m.inputs)).To(gomega.Equal(5))

		b, derr := m.Beat(context.TODO())
		gomega.Expect(derr).To(gomega.Succeed())

		md := b.(*MetricsData)

		// Check if we have data for all expected measurements
		metricList := map[string]string{} // To remove dups
		for _, metric := range(md.Metrics) {
			metricList[metric.Name] = metric.Name
		}

		gomega.Expect(metricList).To(gomega.ConsistOf([]string{"cpu", "disk", "diskio", "mem", "net"}))
	})
})
