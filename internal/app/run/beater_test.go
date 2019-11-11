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

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("beater", func() {
	ginkgo.It("should send heartbeat", func() {
		d, derr := NewDispatcher(testClient, NewWorker(testConfig), 10)
		gomega.Expect(derr).To(gomega.Succeed())

		beater := Beater{testClient, d, "testasset"}

		cur := testHandler.GetNumChecks()
		gomega.Expect(beater.Beat(time.Second)).To(gomega.BeTrue())
		gomega.Expect(testHandler.GetNumChecks()).To(gomega.Equal(cur + 1))
	})

	ginkgo.It("should dispatch received operations", func() {
		d, derr := NewDispatcher(testClient, NewWorker(testConfig), 10)
		gomega.Expect(derr).To(gomega.Succeed())

		beater := Beater{testClient, d, "test-asset"}

		cur := testHandler.GetNumCallbacks()
		gomega.Expect(beater.Beat(time.Second)).To(gomega.BeTrue())

		// Wait until all operations are dealt with
		close(d.opQueue)
		d.opWorkerWaitgroup.Wait()
		close(d.resQueue)
		d.resWorkerWaitgroup.Wait()

		// Three operations, scheduled and succeeded, equals 6 callbacks
		gomega.Expect(testHandler.GetNumCallbacks()).To(gomega.Equal(cur + 6))
	})
})
