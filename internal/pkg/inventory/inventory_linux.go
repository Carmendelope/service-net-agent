/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package inventory

// Gather inventory information

import (
	"fmt"

	"github.com/nalej/derrors"
	"github.com/nalej/grpc-edge-controller-go"
	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/sysinfo"

	"github.com/rs/zerolog/log"
)

// TBD - this package should not be OS-specific; that should be encapsulated
// in nalej/sysinfo. We'll tackle that when we clean that package up and
// introduce Windows compatibility.

type Inventory struct {
	sysinfo.SysInfo
}

func NewInventory() (*Inventory, derrors.Error) {
	log.Debug().Msg("gathering system inventory information")

	i := &Inventory{sysinfo.SysInfo{}}
	i.GetSysInfo()
	log.Debug().Interface("sysinfo", i).Msg("inventory information")

	return i, nil
}

func (i *Inventory) GetRequest() (*grpc_edge_controller_go.AgentJoinRequest) {
	netInfo := []*grpc_inventory_go.NetworkingHardwareInfo{}
	for _, iface := range(i.Network) {
		grpcIface := &grpc_inventory_go.NetworkingHardwareInfo{
			Type: iface.Port,
			LinkCapacity: int64(iface.Speed),
		}
		netInfo = append(netInfo, grpcIface)
	}

	cpuInfo := make([]*grpc_inventory_go.CPUInfo, i.CPU.Cpus)
	for count := uint(0); count  < i.CPU.Cpus; count++ {
		cpuInfo[count] = &grpc_inventory_go.CPUInfo{
			Manufacturer: i.CPU.Vendor,
			Model: i.CPU.Model,
			NumCores: int32(i.CPU.Cores),
		}
	}

	storageInfo := make([]*grpc_inventory_go.StorageHardwareInfo, 0, len(i.Storage))
	for _, s := range(i.Storage) {
		grpcStorage := &grpc_inventory_go.StorageHardwareInfo{
			Type: fmt.Sprintf("%s %s", s.Vendor, s.Model),
			TotalCapacity: int64(s.Size),
		}
		storageInfo = append(storageInfo, grpcStorage)
	}

	// Set labels
	labels := map[string]string{}
	labels["chassis_vendor"] = i.Chassis.Vendor
	labels["chassis_version"] = i.Chassis.Version
	labels["chassis_serial"] = i.Chassis.Serial
	labels["chassis_assettag"] = i.Chassis.AssetTag
	labels["board_name"] = i.Board.Name
	labels["board_vendor"] = i.Board.Vendor
	labels["board_version"] = i.Board.Version
	labels["board_serial"] = i.Board.Serial
	labels["board_assettag"] = i.Board.AssetTag

	// Unset empty ones
	for k, v := range(labels) {
		if v == "" {
			delete(labels, k)
		}
	}

	request := &grpc_edge_controller_go.AgentJoinRequest{
		Labels: labels,
		Os: &grpc_inventory_go.OperatingSystemInfo{
			Name: i.OS.Name,
			Version: i.Kernel.Release,
			Architecture: i.OS.Architecture,
			Class: GetOSClass(),
		},
		Hardware: &grpc_inventory_go.HardwareInfo{
			Cpus: cpuInfo,
			InstalledRam: int64(i.Memory.Size),
			NetInterfaces: netInfo,
		},
		Storage: storageInfo,
	}

	return request
}
