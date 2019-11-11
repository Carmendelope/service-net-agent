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

// Service manager - systemd implementation for linux

package svcmgr

import (
	"go/build"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
)

type Installer struct {
}

func NewInstaller(name, root string) (*Installer, derrors.Error) {
	i := &Installer{}

	return i, nil
}

// Install a service
func (i *Installer) Install(bin string, args []string, desc ...string) derrors.Error {
	log.Warn().Str("bin", bin).Str("os", build.Default.GOOS).Msg("install not implemented")
	return nil
}

// Remove system service
func (i *Installer) Remove() derrors.Error {
	log.Warn().Str("os", build.Default.GOOS).Msg("remove system service not implemented")
	return nil
}

// Enable system service
func (i *Installer) Enable() derrors.Error {
	log.Warn().Str("os", build.Default.GOOS).Msg("enable system service not implemented")
	return nil
}

// Disable system service
func (i *Installer) Disable() derrors.Error {
	log.Warn().Str("os", build.Default.GOOS).Msg("disable system service not implemented")
	return nil
}
