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

// +build !linux,!windows

// Service manager - implementation for unsupported systems

package svcmgr

import (
	"go/build"

	"github.com/nalej/derrors"
)

type Implementation struct {
}

func NewImplementation(string, Runner) (*Implementation, derrors.Error) {
	i := &Implementation{}

	return i, nil
}

func (i *Implementation) Run() derrors.Error {
	return derrors.NewUnimplementedError("starting as service not supported").WithParams(build.Default.GOOS)
}

// On unsupported systems, the system services are never OK
func checkSystem() derrors.Error {
	return derrors.NewUnimplementedError("starting as service not supported").WithParams(build.Default.GOOS)
}

// Stop system service
func Stop(servicename string) derrors.Error {
	return derrors.NewUnimplementedError("stopping service not supported").WithParams(build.Default.GOOS)
}
