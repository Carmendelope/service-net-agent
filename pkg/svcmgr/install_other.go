// +build !linux,!windows

/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

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
	i := &Installer{
	}

	return i, nil
}

// Install a service
func (i *Installer) Install(bin string, args []string, desc ...string) (derrors.Error) {
	log.Warn().Str("bin", bin).Str("os", build.Default.GOOS).Msg("install not implemented")
	return nil
}

// Remove system service
func (i *Installer) Remove() (derrors.Error) {
	log.Warn().Str("os", build.Default.GOOS).Msg("remove system service not implemented")
	return nil
}

// Enable system service
func (i *Installer) Enable() (derrors.Error) {
	log.Warn().Str("os", build.Default.GOOS).Msg("enable system service not implemented")
	return nil
}

// Disable system service
func (i *Installer) Disable() (derrors.Error) {
	log.Warn().Str("os", build.Default.GOOS).Msg("disable system service not implemented")
	return nil
}
