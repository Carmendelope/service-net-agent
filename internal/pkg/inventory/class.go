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

// Mapping of build parameters to OS Class

import (
	"github.com/nalej/grpc-inventory-go"
	"go/build"
)

var osToClass = map[string]grpc_inventory_go.OperatingSystemClass{
	"linux":   grpc_inventory_go.OperatingSystemClass_LINUX,
	"windows": grpc_inventory_go.OperatingSystemClass_WINDOWS,
	"darwin":  grpc_inventory_go.OperatingSystemClass_DARWIN,
}

func GetOSClass() grpc_inventory_go.OperatingSystemClass {
	return osToClass[build.Default.GOOS]
}
