/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package machine_id

import (
	"github.com/coreos/go-systemd/util"
	"github.com/rs/zerolog/log"
)

// See https://www.freedesktop.org/software/systemd/man/sd_id128_get_machine.html
// This ID may be used wherever a unique identifier for the local system is
// needed.
func MachineID() string {
	if !util.IsRunningSystemd() {
		log.Warn().Msg("cannot determine system id - systemd not available")
		return ""
	}

	id, err := util.GetMachineID()
	if err != nil {
		log.Warn().Err(err).Msg("cannot determine system id")
		return ""
	}

	return id
}
