/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package defaults

// Windows Agent defaults

import (
	"os"
)

var (
	Path string = os.Getenv("ProgramFiles") + string(os.PathSeparator) + "nalej"
)
