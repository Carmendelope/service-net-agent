/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package ping

// Ping example plugin

import (
	"fmt"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/spf13/viper"
)

// TBD:
// - example of configuration
// - example of actual functionality in start/stop

type Ping struct {
	config *viper.Viper

	commandMap plugin.CommandMap
}

const (
	pingName plugin.PluginName = "ping"
)

func init() {
	plugin.Register(pingName, NewPing)
}

func NewPing(config *viper.Viper) (plugin.Plugin, derrors.Error) {
	p := &Ping{
		config: config,
	}

	commandMap :=  plugin.CommandMap{
		plugin.CommandName("ping"): p.ping,
	}

	p.commandMap = commandMap

	return p, nil
}

func (p *Ping) StartPlugin() (derrors.Error) {
	return nil
}

func (p *Ping) StopPlugin() {
}

func (p *Ping) GetCommand(cmd plugin.CommandName) plugin.CommandFunc {
	return p.commandMap[cmd]
}

func (p *Ping) ping(params map[string]string) (string, derrors.Error) {
	var msgString string
	msg, found := params["msg"]
	if found {
		msgString = fmt.Sprintf(" %s", msg)
	}

	return "pong" + msgString, nil
}
