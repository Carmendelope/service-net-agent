/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package inventory

// Mapping of build parameters to OS Class

import (
	"go/build"
	"github.com/nalej/grpc-inventory-go"
)

var osToClass = map[string]grpc_inventory_go.OperatingSystemClass{
	"linux": grpc_inventory_go.OperatingSystemClass_LINUX,
	"windows": grpc_inventory_go.OperatingSystemClass_WINDOWS,
	"darwin": grpc_inventory_go.OperatingSystemClass_DARWIN,
}

func GetOSClass() grpc_inventory_go.OperatingSystemClass {
	return osToClass[build.Default.GOOS]
}
