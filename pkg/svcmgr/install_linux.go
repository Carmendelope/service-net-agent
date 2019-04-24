/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Service manager - systemd implementation for linux

package svcmgr

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
)

type Installer struct {
	name string
	root string
}

func NewInstaller(name, root string) (*Installer, derrors.Error) {
	derr := checkSystem() // Won't install if we're not running systemd
	if derr != nil {
		return nil, derr
	}

	i := &Installer{
		name: name,
		root: root,
	}

	return i, nil
}

// Install a service
func (i *Installer) Install(bin string, args []string, desc ...string) (derrors.Error) {
	log.Debug().Str("name", i.name).Str("bin", bin).Msg("installing service")

	// Check if exists and executable
	binPath, err := exec.LookPath(bin)
	if err != nil {
		return derrors.NewUnavailableError("service executable not found", err).WithParams(bin)
	}
	absBinPath, err := filepath.Abs(binPath)
	if err != nil {
		return derrors.NewUnavailableError("service executable not found", err).WithParams(bin)
	}

	reader, derr := createUnitFile(strings.Join(desc, " "), absBinPath, args)
	if derr != nil {
		return derr
	}

	// Determine unit filename
	filename, derr := getUnitFilename(i.name, i.root)
	if derr != nil {
		return derr
	}

	unitPath := filepath.Dir(filename)
	err = os.MkdirAll(unitPath, 0755)
	if err != nil {
		return derrors.NewPermissionDeniedError("unable to create unit path", err).WithParams(unitPath)
	}

	// Write file - just overwrite if already exists
	unitFile, err := os.Create(filename)
	if err != nil {
		return derrors.NewPermissionDeniedError("unable to create unit file", err).WithParams(filename)
	}
	defer unitFile.Close()

	_, err = io.Copy(unitFile, reader)
	if err != nil {
		return derrors.NewInternalError("error writing unit file", err).WithParams(filename)
	}

	return nil
}

// Enable system service
func (i *Installer) Enable() (derrors.Error) {
	log.Debug().Str("name", i.name).Msg("enabling system service")

	// Determine unit filename
	filename, derr := getUnitFilename(i.name, i.root)
	if derr != nil {
		return derr
	}

	// Check if exists
	_, err := os.Stat(filename)
	if err != nil {
		return derrors.NewNotFoundError("system service file not found").WithParams(filename)
	}

	return enableUnit(filename)
}
