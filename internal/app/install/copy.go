/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package install

// Copy files

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

func copyFile(dest, src string) error {
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
	destFile, err := os.OpenFile(dest, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, srcStat.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
