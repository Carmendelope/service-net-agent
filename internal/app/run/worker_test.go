/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
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
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "start", params)).To(gomega.BeZero())

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
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "start", nil)).To(gomega.BeZero())

		// And stop
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "stop", nil)).To(gomega.BeZero())
	})

	ginkgo.It("should execute a plugin command", func() {
		worker := NewWorker(testConfig)
		gomega.Expect(worker.Execute(context.Background(), testPlugin, "start", nil)).To(gomega.BeZero())

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
