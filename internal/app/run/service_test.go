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
	ginkgo.Context("heartbeat", func(){
		ginkgo.It("should send heartbeat", func() {
			d, derr := NewDispatcher(testClient, NewWorker(testConfig), 10)
			gomega.Expect(derr).To(gomega.Succeed())

			cur := testHandler.GetNumChecks()
			gomega.Expect(heartbeat(testClient, d, "testasset")).To(gomega.BeTrue())
			gomega.Expect(testHandler.GetNumChecks()).To(gomega.Equal(cur + 1))
		})

		ginkgo.It("should dispatch received operations", func() {
			d, derr := NewDispatcher(testClient, NewWorker(testConfig), 10)
			gomega.Expect(derr).To(gomega.Succeed())

			cur := testHandler.GetNumCallbacks()
			gomega.Expect(heartbeat(testClient, d, "test-asset")).To(gomega.BeTrue())

			// Wait until all operations are dealt with
			close(d.opQueue)
			d.opWorkerWaitgroup.Wait()
			close(d.resQueue)
			d.resWorkerWaitgroup.Wait()

			// Three operations, scheduled and succeeded, equals 6 callbacks
			gomega.Expect(testHandler.GetNumCallbacks()).To(gomega.Equal(cur + 6))
		})
	})

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
