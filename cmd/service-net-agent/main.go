/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Service Net Agent main executable

package main

import (
	"github.com/nalej/service-net-agent/cmd/service-net-agent/commands"
	"github.com/nalej/service-net-agent/version"
)

var MainVersion string

var MainCommit string

func main() {
	version.AppVersion = MainVersion
	version.Commit = MainCommit
	commands.Execute()
}
