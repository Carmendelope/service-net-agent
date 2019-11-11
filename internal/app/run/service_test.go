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

		gomega.Expect(testHandler.GetNumChecks()).To(gomega.BeNumerically(">=", cur+2))
	})
	ginkgo.It("should start, run and disable and stop", func() {
		s := Service{
			Config: testConfig,
			Client: testClient,
		}

		testConfig.Set("agent.asset_id", "test-asset")
		testConfig.Set("agent.interval", "0.001s")
		testConfig.Set("agent.opqueue_len", 10)
		testConfig.Set("agent.shutdown_timeout", "10s")
		testConfig.Set("agent.comm_timeout", "10s")

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

		s.Disable()

		// Should die but not stop
		for {
			alive, derr := s.Alive()
			gomega.Expect(derr).To(gomega.Succeed())
			if !alive {
				break
			}
			time.Sleep(time.Second / 1000) // One heartbeat period
		}

		s.Stop()
		derr = <-errChan // wait until done
		gomega.Expect(derr).To(gomega.Succeed())
	})
})
