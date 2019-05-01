/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Service manager - utility functions

package svcmgr

import (
	"os/exec"
	"path/filepath"

	"github.com/nalej/derrors"
)

func fullExecPath(bin string) (string, derrors.Error) {
	// Check if exists and executable
	binPath, err := exec.LookPath(bin)
	if err != nil {
		return "", derrors.NewUnavailableError("service executable not found", err).WithParams(bin)
	}
	absBinPath, err := filepath.Abs(binPath)
	if err != nil {
		return "", derrors.NewUnavailableError("service executable not found", err).WithParams(bin)
	}

	return absBinPath, nil
}
