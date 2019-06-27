/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func TestHandlerPackage(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "internal/pkg/agentplugin/metrics package suite")
}

var (
	testName string
	testTags map[string]string
	testFields map[string]interface{}

	testMetric *Metric
)

var _ = ginkgo.BeforeSuite(func() {
	testName = "test"

	testTags = map[string]string{
		"testtag1": "testval1",
		"testtag2": "testval2",
	}

	testFields = map[string]interface{}{
		"field1": uint64(12345),
		"field2": float64(1.234),
	}

	testMetric = &Metric{
		Name: testName,
		Tags: testTags,
		Fields: map[string]uint64{
			"field1": 12345,
			"field2": 1,
		},
	}
})
