/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package test

import (
	"context"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-controller-go"
	"github.com/nalej/infra-net-plugin"
	_ "github.com/nalej/infra-net-plugin/ping"

	"github.com/nalej/service-net-agent/internal/pkg/agentplugin"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/spf13/viper"
)

var heartbeatDescriptor = plugin.PluginDescriptor{
	Name: "heartbeat",
	Description: "test agent heartbeat plugin",
	NewFunc: NewHeartbeat,
}

type Heartbeat struct {
	agentplugin.BaseAgentPlugin
	config *viper.Viper
}

func init() {
	plugin.Register(&heartbeatDescriptor)
}

func NewHeartbeat(config *viper.Viper) (plugin.Plugin, derrors.Error) {
	hb := &Heartbeat{
		config: config,
	}

	return hb, nil
}

func (hb *Heartbeat) GetPluginDescriptor() *plugin.PluginDescriptor {
	return &heartbeatDescriptor
}

func (hb *Heartbeat) Beat(context.Context) (agentplugin.PluginHeartbeatData, derrors.Error) {
	wait := hb.config.GetDuration("beatsleep")
	if wait > 0 {
		time.Sleep(wait)
	}
	return &HeartbeatData{}, nil
}

type HeartbeatData struct {}

func (d *HeartbeatData) ToGRPC() *grpc_edge_controller_go.PluginData {
	data := &grpc_edge_controller_go.PluginData{
		// TODO: Make plugin data structures in gRPC for testing
	}

	return data
}

var _ = ginkgo.Describe("heartbeat", func() {
	ginkgo.Context("Registry", func() {
		ginkgo.AfterEach(func() {
			plugin.StopAll()
		})
		ginkgo.It("should be able to collect heartbeat data from running plugins", func() {
			err := plugin.StartPlugin("heartbeat", nil)
			gomega.Expect(err).To(gomega.Succeed())

			data, errlist := agentplugin.CollectHeartbeatData(context.Background())
			gomega.Expect(errlist).To(gomega.BeEmpty())
			gomega.Expect(data).To(gomega.HaveLen(1))
		})
		ginkgo.It("should be able to timeout on heartbeat data collection", func() {
			conf := viper.New()
			conf.Set("beatsleep", "10s")
			err := plugin.StartPlugin("heartbeat", conf)
			gomega.Expect(err).To(gomega.Succeed())

			ctx, _ := context.WithDeadline(context.Background(), time.Now())
			data, errlist := agentplugin.CollectHeartbeatData(ctx)
			gomega.Expect(errlist).To(gomega.HaveLen(1))
			gomega.Expect(errlist["heartbeat"].Type()).To(gomega.Equal(derrors.DeadlineExceeded))
			gomega.Expect(data).To(gomega.BeEmpty())
		})
		ginkgo.It("should not collect heartbeat data from stopped plugins", func() {
			data, err := agentplugin.CollectHeartbeatData(context.Background())
			gomega.Expect(err).To(gomega.BeEmpty())
			gomega.Expect(data).To(gomega.BeEmpty())
		})
		ginkgo.It("should not collect heartbeat data from non-agent plugins", func() {
			err := plugin.StartPlugin("ping", nil)
			gomega.Expect(err).To(gomega.Succeed())

			data, errlist := agentplugin.CollectHeartbeatData(context.Background())
			gomega.Expect(errlist).To(gomega.BeEmpty())
			gomega.Expect(data).To(gomega.BeEmpty())
		})
	})
})

