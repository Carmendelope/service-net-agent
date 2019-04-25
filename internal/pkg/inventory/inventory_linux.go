/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package inventory

// Gather inventory information

import (
	"fmt"

	"github.com/nalej/derrors"
	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/sysinfo"

	"github.com/rs/zerolog/log"
)

func NewInventory() (*Inventory, derrors.Error) {
	log.Debug().Msg("gathering system inventory information")

	info := sysinfo.SysInfo{}
	info.GetSysInfo()
	log.Debug().Interface("sysinfo", info).Msg("inventory information")

	netInfo := []*grpc_inventory_go.NetworkingHardwareInfo{}
	for _, iface := range(info.Network) {
		grpcIface := &grpc_inventory_go.NetworkingHardwareInfo{
			Type: iface.Port,
			LinkCapacity: int64(iface.Speed),
		}
		netInfo = append(netInfo, grpcIface)
	}

	i := &Inventory{
		Os: &grpc_inventory_go.OperatingSystemInfo{
			Name: info.OS.Name,
			Version: info.OS.Release,

		},
		Hardware: &grpc_inventory_go.HardwareInfo{
			Cpus: []*grpc_inventory_go.CPUInfo{
				&grpc_inventory_go.CPUInfo{
					Manufacturer: info.CPU.Vendor,
					Model: info.CPU.Model,
					Architecture: info.OS.Architecture,
					NumCores: int32(info.CPU.Threads),
				},
			},
			InstalledRam: int64(info.Memory.Size),
			NetInterfaces: netInfo,

		},
		Storage: &grpc_inventory_go.StorageHardwareInfo{
			Type: fmt.Sprintf("%s %s", info.Storage[0].Vendor, info.Storage[0].Model),
			TotalCapacity: int64(info.Storage[0].Size),
		},
	}

	return i, nil
}
