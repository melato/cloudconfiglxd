package lxdcloudinit

import (
	"fmt"
	"os"

	lxd "github.com/lxc/lxd/client"
	"melato.org/lxdcloudinit/lxdutil"
)

type App struct {
	lxdutil.LxdClient
	Instance string `name:"i" usage:"LXD instance to configure"`
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
	t.server = server
	return nil
}

func (t *App) Apply(files ...string) error {
	configurer := NewConfigurer(t.server)
	configurer.Log = os.Stdout
	return configurer.ApplyConfigFiles(t.Instance, files...)
}
