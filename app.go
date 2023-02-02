package lxdcloudinit

import (
	"fmt"
	"os"

	lxd "github.com/lxc/lxd/client"
	"melato.org/cloudinit"
	"melato.org/cloudinit/ostype"
	"melato.org/lxdcloudinit/lxdutil"
)

type App struct {
	lxdutil.LxdClient
	Instance string `name:"i" usage:"LXD instance to configure"`
	OS       string `name:"ostype" usage:"OS type"`
	os       cloudinit.OSType
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
	base := NewInstanceConfigurer(t.server, t.Instance)
	base.Log = os.Stdout
	configurer := cloudinit.NewConfigurer(base)
	configurer.OS = t.os
	configurer.Log = os.Stdout
	return configurer.ApplyConfigFiles(files...)
}
