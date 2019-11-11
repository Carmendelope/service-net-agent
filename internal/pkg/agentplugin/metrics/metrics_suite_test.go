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
	testName   string
	testTags   map[string]string
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
