/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package ping

// Ping example plugin

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/spf13/viper"
)

// TBD:
// - example of configuration
// - example of actual functionality in start/stop

var pingDescriptor = plugin.PluginDescriptor{
	Name: "ping",
	Description: "example ping plugin",
	NewFunc: NewPing,
}

type Ping struct {
	config *viper.Viper

	commandMap plugin.CommandFuncMap
}

func init() {
	pingCommand := plugin.CommandDescriptor{
		Name: "ping",
		Description: "echo command",
	}
	pingCommand.AddParam(plugin.ParamDescriptor{
		Name: "msg",
		Description: "message to be echoed",
		Required: false,
	})
	pingCommand.AddParam(plugin.ParamDescriptor{
		Name: "sleep",
		Description: "time in seconds to wait before echoing message",
		Required: false,
		Default: "0",
	})

	pingDescriptor.AddCommand(pingCommand)

	plugin.Register(&pingDescriptor)
}

func NewPing(config *viper.Viper) (plugin.Plugin, derrors.Error) {
	p := &Ping{
		config: config,
	}

	commandMap :=  plugin.CommandFuncMap{
		"ping": p.ping,
	}

	p.commandMap = commandMap

	return p, nil
}

func (p *Ping) StartPlugin() (derrors.Error) {
	return nil
}

func (p *Ping) StopPlugin() {
}

func (p *Ping) GetCommandFunc(cmd plugin.CommandName) plugin.CommandFunc {
	return p.commandMap[cmd]
}

func (p *Ping) GetPluginDescriptor() *plugin.PluginDescriptor {
	return &pingDescriptor
}

func (p *Ping) Beat(context.Context) (plugin.PluginHeartbeatData, derrors.Error) {
	return &PingData{}, nil
}

func (p *Ping) ping(ctx context.Context, params map[string]string) (string, derrors.Error) {
	var msgString string
	msg, found := params["msg"]
	if found {
		msgString = fmt.Sprintf(" %s", msg)
	}
	sleep, _ := strconv.Atoi(params["sleep"])
	if sleep > 0 {
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	return "pong" + msgString, nil
}
