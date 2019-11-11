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
	"github.com/nalej/grpc-inventory-go"
	"time"

	"github.com/nalej/grpc-inventory-manager-go"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("dispatch", func() {

	var testRequest = &grpc_inventory_manager_go.AgentOpRequest{
		OrganizationId:   "testorg",
		EdgeControllerId: "testec",
		AssetId:          "testasset",
		Plugin:           testPlugin,
		Operation:        "start",
		OperationId:      "testop",
	}

	ginkgo.Context("opWorker", func() {
		ginkgo.It("should execute operations from queue", func() {
			// We're sending a request on opQueue and expecting a response on resQueue
			d := &Dispatcher{
				worker:   NewWorker(testConfig),
				opQueue:  make(chan *grpc_inventory_manager_go.AgentOpRequest, 1),
				resQueue: make(chan *grpc_inventory_manager_go.AgentOpResponse, 1),
			}

			// Put request on operation queue
			d.opQueue <- testRequest
			close(d.opQueue) // Can read request, but worker will bail out after

			d.opWorkerWaitgroup.Add(1)
			d.opWorker(context.Background())

			response := <-d.resQueue
			gomega.Expect(response.GetOrganizationId()).To(gomega.Equal("testorg"))
			gomega.Expect(response.GetEdgeControllerId()).To(gomega.Equal("testec"))
			gomega.Expect(response.GetAssetId()).To(gomega.Equal("testasset"))
			gomega.Expect(response.GetOperationId()).To(gomega.Equal("testop"))
			gomega.Expect(time.Unix(response.GetTimestamp(), 0)).To(gomega.BeTemporally("~", time.Now(), time.Second))
			gomega.Expect(response.GetStatus()).To(gomega.Equal(grpc_inventory_go.OpStatus_SUCCESS))
		})
	})

	ginkgo.Context("resWorker", func() {
		ginkgo.It("should send responses from queue", func() {
			d := &Dispatcher{
				client:   testClient,
				resQueue: make(chan *grpc_inventory_manager_go.AgentOpResponse, 1),
			}

			d.respond(testRequest, grpc_inventory_go.OpStatus_SUCCESS, "")
			close(d.resQueue)

			cur := testHandler.GetNumCallbacks()
			d.resWorkerWaitgroup.Add(1)
			d.resWorker(context.Background())

			// Response should have been processed
			gomega.Expect(len(d.resQueue)).To(gomega.BeZero())
			gomega.Expect(testHandler.GetNumCallbacks()).To(gomega.Equal(cur + 1))
		})
	})

	ginkgo.It("should dispatch operations", func() {
		d, derr := NewDispatcher(testClient, NewWorker(testConfig), 10)
		gomega.Expect(derr).To(gomega.Succeed())

		cur := testHandler.GetNumCallbacks()

		derr = d.Dispatch(testRequest)
		gomega.Expect(derr).To(gomega.Succeed())

		derr = d.Stop(time.Second)
		gomega.Expect(derr).To(gomega.Succeed())
		gomega.Expect(len(d.resQueue)).To(gomega.BeZero())
		gomega.Expect(len(d.opQueue)).To(gomega.BeZero())
		// Two callbacks: scheduled and success
		gomega.Expect(testHandler.GetNumCallbacks()).To(gomega.Equal(cur + 2))
	})

	ginkgo.It("should cancel queued operations when stopped", func() {
		_, cancelOpWorker := context.WithCancel(context.Background())

		d := &Dispatcher{
			opQueue:        make(chan *grpc_inventory_manager_go.AgentOpRequest, 1),
			resQueue:       make(chan *grpc_inventory_manager_go.AgentOpResponse, 1),
			cancelOpWorker: cancelOpWorker,
		}

		d.opQueue <- testRequest

		derr := d.Stop(time.Second)
		gomega.Expect(derr).To(gomega.Succeed())

		response := <-d.resQueue
		gomega.Expect(response.GetStatus()).To(gomega.Equal(grpc_inventory_go.OpStatus_FAIL))
		gomega.Expect(response.GetInfo()).To(gomega.ContainSubstring("stopped"))
	})

	ginkgo.It("should not accept operations when the queue is full", func() {
		d := &Dispatcher{
			opQueue:  make(chan *grpc_inventory_manager_go.AgentOpRequest, 0),
			resQueue: make(chan *grpc_inventory_manager_go.AgentOpResponse, 1),
		}

		derr := d.Dispatch(testRequest)
		gomega.Expect(derr).To(gomega.Succeed())

		response := <-d.resQueue
		gomega.Expect(response.GetStatus()).To(gomega.Equal(grpc_inventory_go.OpStatus_FAIL))
		gomega.Expect(response.GetInfo()).To(gomega.ContainSubstring("full"))
	})
})
