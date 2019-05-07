/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package machine_id

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/denisbrodbeck/machineid"
	"github.com/rs/zerolog/log"
)

// In untrusted environments it is recommended to derive an
// application specific ID from this machine ID, in an irreversable
// (cryptographically secure) way. To make this easy AppSpecificMachineId()
// is provided

// Similar to sd_id128_get_machine_app_specific
func AppSpecificMachineID(appId string) string {
	// machineid is a cross-platform library to retrieve a unique
	// id for a machine. On Linux, it uses code similar to what's
	// found in coreos/go-systemd.
	machineID, err := machineid.ID()
	if err != nil || machineID == "" {
		log.Warn().Err(err).Msg("unable to get machine id")
		return ""
	}

	h := hmac.New(sha256.New, []byte(machineID))
	h.Write([]byte(appId))

	sha := hex.EncodeToString(h.Sum(nil))
	// Create dashed uuid
	id := fmt.Sprintf("%s-%s-%s-%s-%s", sha[0:8], sha[8:12], sha[12:16], sha[16:20], sha[20:32])
	return id
}
