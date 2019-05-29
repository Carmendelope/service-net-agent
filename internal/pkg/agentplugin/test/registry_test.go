/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package test

import (
	"context"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/pkg/plugin"
	"github.com/nalej/service-net-agent/internal/pkg/agentplugin"
	_ "github.com/nalej/service-net-agent/internal/pkg/agentplugin/ping"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/spf13/viper"
)

var _ = ginkgo.Describe("plugin", func() {

	var expectedPingDescriptor = &plugin.PluginDescriptor{
		Name: "ping",
		Description: "example ping plugin",
		Commands: plugin.CommandMap{
			"ping": plugin.CommandDescriptor{
				Name: "ping",
				Description: "echo command",
				Params: plugin.ParamMap{
					"msg": plugin.ParamDescriptor{
						Name: "msg",
						Description: "message to be echoed",
					},
					"sleep": plugin.ParamDescriptor{
						Name: "sleep",
						Description: "time in seconds to wait before echoing message",
						Default: "0",
					},
				},
			},
		},
	}

	ginkgo.BeforeSuite(func() {

	})

	ginkgo.Context("Registry", func() {
	// Importing ping already executes a lot of functionality. We'll just check
	// if the result is what we expect
		ginkgo.AfterEach(func() {
			plugin.StopAll()
		})

		ginkgo.It("should register and list plugins", func() {
			plugins := plugin.ListPlugins()

			// Hack: we can't seem to compare function pointers
			newFunc := plugins["ping"].NewFunc
			plugins["ping"].NewFunc = nil

			// This will test the registration and storage of plugins,
			// as well as the functions used to build the descriptor.
			expected := plugin.RegistryEntryMap{
				"ping": plugin.RegistryEntry{
					PluginDescriptor: expectedPingDescriptor,
					Running: false,
				},
			}

			gomega.Expect(plugins).To(gomega.Equal(expected))

			// ;(
			plugins["ping"].NewFunc = newFunc
		})
		ginkgo.It("should not be able to register a plugin twice", func() {
			err := plugin.Register(expectedPingDescriptor)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should start and stop a plugin", func() {
			err := plugin.StartPlugin("ping", nil)
			gomega.Expect(err).To(gomega.Succeed())

			err = plugin.StopPlugin("ping")
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should not start a non-existing plugin", func() {
			err := plugin.StartPlugin("invalid", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should not start a plugin twice", func() {
			err := plugin.StartPlugin("ping", nil)
			gomega.Expect(err).To(gomega.Succeed())
			err = plugin.StartPlugin("ping", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should not stop a stopped plugin", func() {
			err := plugin.StopPlugin("ping")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should execute a command on a running plugin", func() {
			err := plugin.StartPlugin("ping", nil)
			gomega.Expect(err).To(gomega.Succeed())

			gomega.Expect(plugin.ExecuteCommand(context.Background(), "ping", "ping", nil)).To(gomega.Equal("pong"))
		})
		ginkgo.It("should not execute a command on a stopped plugin", func() {
			_, err := plugin.ExecuteCommand(context.Background(), "ping", "ping", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should not execute a command on a non-existing plugin", func() {
			_, err := plugin.ExecuteCommand(context.Background(), "invalid", "ping", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())

		})
		ginkgo.It("should not execute a non-existing command", func() {
			err := plugin.StartPlugin("ping", nil)
			gomega.Expect(err).To(gomega.Succeed())

			_, err = plugin.ExecuteCommand(context.Background(), "ping", "invalid", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should be able to collect heartbeat data from running plugins", func() {
			err := plugin.StartPlugin("ping", nil)
			gomega.Expect(err).To(gomega.Succeed())

			data, errlist := plugin.CollectHeartbeatData(context.Background())
			gomega.Expect(errlist).To(gomega.BeEmpty())
			gomega.Expect(data).To(gomega.HaveLen(1))
		})
		ginkgo.It("should be able to timeout on heartbeat data collection", func() {
			conf := viper.New()
			conf.Set("beatsleep", "10s")
			err := plugin.StartPlugin("ping", conf)
			gomega.Expect(err).To(gomega.Succeed())

			ctx, _ := context.WithDeadline(context.Background(), time.Now())
			data, errlist := plugin.CollectHeartbeatData(ctx)
			gomega.Expect(errlist).To(gomega.HaveLen(1))
			gomega.Expect(errlist["ping"].Type()).To(gomega.Equal(derrors.DeadlineExceeded))
			gomega.Expect(data).To(gomega.BeEmpty())
		})
		ginkgo.It("should not collect heartbeat data from stopped plugins", func() {
			data, err := plugin.CollectHeartbeatData(context.Background())
			gomega.Expect(err).To(gomega.BeEmpty())
			gomega.Expect(data).To(gomega.BeEmpty())
		})
	})
})
