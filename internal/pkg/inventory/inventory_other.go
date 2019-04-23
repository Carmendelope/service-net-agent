// +build !linux

/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package inventory

// Gather inventory information

import (
	"go/build"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"
)

func NewInventory() (*Inventory, derrors.Error) {
	log.Warn().Str("os", build.Default.GOOS).Msg("gathering inventory not supported")

	i := &Inventory{}
	return i, nil
}
