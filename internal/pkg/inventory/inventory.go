/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package inventory

// Gather inventory information

import (
	"github.com/nalej/grpc-inventory-go"
)

type Inventory struct {
	// OS contains Operating System information.
	Os *grpc_inventory_go.OperatingSystemInfo
	// Hardware information.
	Hardware *grpc_inventory_go.HardwareInfo
	// Storage information.
	Storage *grpc_inventory_go.StorageHardwareInfo
}

// TBD - this package should not be OS-specific; that should be encapsulated
// in nalej/sysinfo. We'll tackle that when we clean that package up and
// introduce Windows compatibility.
