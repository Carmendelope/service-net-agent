/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
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
