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

package inventory

import (
	"github.com/nalej/grpc-edge-controller-go"
	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/sysinfo"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("inventory", func() {

	var inventory *Inventory

	ginkgo.BeforeSuite(func() {
		inventory = &Inventory{sysinfo.NewFakeSysInfo()}

		// To test empty labels
		inventory.Board.AssetTag = ""
	})

	ginkgo.It("should correctly translate inventory to request", func() {
		expected := &grpc_edge_controller_go.AgentJoinRequest{
			Labels: map[string]string{
				"board_vendor":     "Test board vendor",
				"board_version":    "v1.0",
				"chassis_vendor":   "Test chassis vendor",
				"chassis_serial":   "serial888",
				"chassis_assettag": "asset888",
				"chassis_version":  "v1.0",
				"board_name":       "Test board",
				"board_serial":     "serial12345",
			},
			Os: &grpc_inventory_go.OperatingSystemInfo{
				Name:         "Test OS",
				Version:      "1.2.3-abc",
				Class:        GetOSClass(),
				Architecture: "aarch32",
			},
			Hardware: &grpc_inventory_go.HardwareInfo{
				Cpus: []*grpc_inventory_go.CPUInfo{
					&grpc_inventory_go.CPUInfo{
						Manufacturer: "Test cpu vendor",
						Model:        "Test cpu model",
						NumCores:     8,
					},
					&grpc_inventory_go.CPUInfo{
						Manufacturer: "Test cpu vendor",
						Model:        "Test cpu model",
						NumCores:     8,
					},
					&grpc_inventory_go.CPUInfo{
						Manufacturer: "Test cpu vendor",
						Model:        "Test cpu model",
						NumCores:     8,
					},
					&grpc_inventory_go.CPUInfo{
						Manufacturer: "Test cpu vendor",
						Model:        "Test cpu model",
						NumCores:     8,
					},
				},
				InstalledRam: 10240,
				NetInterfaces: []*grpc_inventory_go.NetworkingHardwareInfo{
					&grpc_inventory_go.NetworkingHardwareInfo{
						Type:         "tp",
						LinkCapacity: 10000,
					},
					&grpc_inventory_go.NetworkingHardwareInfo{
						Type:         "tp",
						LinkCapacity: 1000,
					},
				},
			},
			Storage: []*grpc_inventory_go.StorageHardwareInfo{
				&grpc_inventory_go.StorageHardwareInfo{
					Type:          "Test storage vendor 0 Test storage model 0",
					TotalCapacity: 10000000,
				},
				&grpc_inventory_go.StorageHardwareInfo{
					Type:          "Test storage vendor 1 Test storage model 1",
					TotalCapacity: 20000000,
				},
			},
		}
		request := inventory.GetRequest()
		gomega.Expect(request).To(gomega.Equal(expected))
	})
})
