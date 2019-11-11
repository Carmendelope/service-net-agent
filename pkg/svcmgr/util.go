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
