/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/nalej/grpc-edge-controller-go"
	"github.com/nalej/grpc-utils/pkg/test"

	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/config"
	"github.com/nalej/service-net-agent/internal/pkg/ec-stub"
	"github.com/nalej/service-net-agent/pkg/plugin"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestHandlerPackage(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "internal/app/run package suite")
}

const (
	testPlugin = "ping"
)

var (
	testConfig *config.Config
	testClient *client.AgentClient

	testPath string
	testConfigFile string

	testListener *bufconn.Listener
	testHandler *ec_stub.Handler
	testServer *grpc.Server
)

var _ = ginkgo.BeforeSuite(func() {
	// Config file for testing plugin persistence
	var err error
	testPath, err = ioutil.TempDir("", "testdata")
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(testPath).To(gomega.BeADirectory())
	testConfigFile = filepath.Join(testPath, "testconfig.yml")

	// Create stub Edge Controller and client
	testListener = test.GetDefaultListener()
	testServer = grpc.NewServer()
	conn, err := test.GetConn(*testListener)
	gomega.Expect(err).To(gomega.Succeed())

	testHandler = ec_stub.NewHandler()
	grpc_edge_controller_go.RegisterAgentServer(testServer, testHandler)
	test.LaunchServer(testServer, testListener)

	testClient = client.NewFakeAgentClient(conn)
})

var _ = ginkgo.AfterSuite(func() {
	// Delete temporary files
	err := os.RemoveAll(testPath)
	gomega.Expect(err).To(gomega.Succeed())

	// Stop server
	testServer.Stop()
	testListener.Close()
})

var _ = ginkgo.BeforeEach(func() {
	// Create new config instance
	testConfig = config.NewConfig()
	testConfig.Path = testPath
	testConfig.ConfigFile = testConfigFile
})

var _ = ginkgo.AfterEach(func() {
	// Stop all running plugins
	plugin.StopAll()

	// Delete config file and instance
	os.Remove(testConfigFile)
	testConfig = nil
})
