/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package install

// Copy files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func copyFile(src, destDir string) error {
	// Destination is destination dir plus filename
	dest := filepath.Join(destDir, filepath.Base(src))

	log.Debug().Str("src", src).Str("dest", dest).Msg("Copying file")

	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return  err
	}
	defer srcFile.Close()

	// Create file with the same permissions as source
	destFile, err := os.OpenFile(dest, os.O_WRONLY | os.O_CREATE, srcStat.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
