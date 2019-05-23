/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

import (
	"time"

	"github.com/nalej/derrors"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("service", func() {
	ginkgo.It("should start, run and stop", func() {
		s := Service{
			Config: testConfig,
			Client: testClient,
		}

		testConfig.Set("agent.asset_id", "test-asset")
		testConfig.Set("agent.interval", "0.001s")
		testConfig.Set("agent.opqueue_len", 10)
		testConfig.Set("agent.shutdown_timeout", "10s")
		testConfig.Set("agent.comm_timeout", "10s")

		cur := testHandler.GetNumChecks()

		errChan := make(chan derrors.Error, 1)
		derr := s.Start(errChan)
		gomega.Expect(derr).To(gomega.Succeed())

		// Wait for at least one heartbeat
		for {
			alive, derr := s.Alive()
			gomega.Expect(derr).To(gomega.Succeed())
			if alive {
				break
			}
			time.Sleep(time.Second / 1000) // One heartbeat period
		}

		s.Stop()
		derr = <-errChan // wait until done
		gomega.Expect(derr).To(gomega.Succeed())

		gomega.Expect(testHandler.GetNumChecks()).To(gomega.BeNumerically(">=", cur + 2))
	})
})
