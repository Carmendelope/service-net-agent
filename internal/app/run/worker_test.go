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

package run

import (
	"context"

	"github.com/nalej/service-net-agent/internal/pkg/config"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("worker", func() {
	ginkgo.It("should start a plugin and write config", func() {
		params := map[string]string{
			"testparam1": "value1",
			"testparam2": "value2",
		}

		worker := NewWorker(testConfig.GetSubConfig("test"))
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "start", params)).To(gomega.Equal("ping enabled"))

		// This above should have written the config. Check that.
		c2 := config.NewConfig()
		c2.ConfigFile = testConfigFile
		gomega.Expect(c2.Read()).To(gomega.Succeed())

		// Worker adds the "enabled" param, which is important because
		// that is checked when starting the plugin again
		params["enabled"] = "true"
		gomega.Expect(c2.GetStringMapString("test." + testPlugin)).To(gomega.Equal(params))
	})

	ginkgo.It("should stop a plugin", func() {
		worker := NewWorker(testConfig)
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "start", nil)).To(gomega.Equal("ping enabled"))

		// And stop
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "stop", nil)).To(gomega.Equal("ping disabled"))
	})

	ginkgo.It("should execute a plugin command", func() {
		worker := NewWorker(testConfig)
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "start", nil)).To(gomega.Equal("ping enabled"))

		params := map[string]string{
			"msg": "test",
		}
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "ping", params)).To(gomega.Equal("pong test"))
	})

	ginkgo.It("should not write config file when plugin doesn't start", func() {
		worker := NewWorker(testConfig.GetSubConfig("test"))
		_, err := worker.Execute(context.Background(), "invalid", "start", nil)
		gomega.Expect(err).To(gomega.HaveOccurred())

		gomega.Expect(testConfigFile).ToNot(gomega.BeAnExistingFile())
	})
})
