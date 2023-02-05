package main

import (
	_ "embed"
	"fmt"
	"os"

	lxd "github.com/lxc/lxd/client"
	"melato.org/cloudconfig"
	"melato.org/cloudconfig/ostype"
	"melato.org/cloudconfiglxd"
	"melato.org/command"
	"melato.org/command/usage"
	"melato.org/lxdclient"
)

//go:embed version
var version string

//go:embed usage.yaml
var usageData []byte

type App struct {
	lxdclient.LxdClient
	Instance string `name:"i" usage:"LXD instance to configure"`
	OS       string `name:"ostype" usage:"OS type"`
	os       cloudconfig.OSType
	server   lxd.InstanceServer
}

func (t *App) Configured() error {
	if t.Instance == "" {
		return fmt.Errorf("missing instance")
	}
	server, err := t.CurrentServer()
	if err != nil {
		return err
	}
	switch t.OS {
	case "":
	case "alpine":
		t.os = &ostype.Alpine{}
	case "debian":
		t.os = &ostype.Debian{}
	default:
		return fmt.Errorf("unknown OS type: %s", t.OS)
	}
	t.server = server
	return nil
}

func (t *App) Apply(files ...string) error {
	base := cloudconfiglxd.NewInstanceConfigurer(t.server, t.Instance)
	base.Log = os.Stdout
	configurer := cloudconfig.NewConfigurer(base)
	configurer.OS = t.os
	configurer.Log = os.Stdout
	return configurer.ApplyConfigFiles(files...)
}

func main() {
	cmd := &command.SimpleCommand{}
	var app App
	cmd.Command("apply").Flags(&app).RunFunc(app.Apply)
	cmd.Command("version").RunFunc(func() { fmt.Println(version) })

	usage.Apply(cmd, usageData)
	command.Main(cmd)
}
