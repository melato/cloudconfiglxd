package lxdcloudinit

import (
	"fmt"
	"io"
	"os"
	"strings"

	lxd "github.com/lxc/lxd/client"
	"melato.org/cloudinit"
	"melato.org/lxdcloudinit/lxdutil"
	"melato.org/yaml"
)

var Trace bool

type Configurer struct {
	Server lxd.InstanceServer
	Log    io.Writer
}

func NewConfigurer(server lxd.InstanceServer) *Configurer {
	return &Configurer{Server: server}
}

func (t *Configurer) log(format string, args ...any) {
	if t.Log != nil {
		fmt.Fprintf(t.Log, format, args...)
	}
}

func (t *Configurer) WriteFile(instance string, f *cloudinit.File) error {
	t.log("write file: %s\n", f.Path)
	var args lxd.InstanceFileArgs
	args.Content = strings.NewReader(f.Content)
	err := t.Server.CreateInstanceFile(instance, f.Path, args)
	if err != nil {
		return lxdutil.AnnotateLXDError(f.Path, err)
	}
	return nil
}

func (t *Configurer) Apply(instance string, config *cloudinit.Config) error {
	for _, f := range config.Files {
		err := t.WriteFile(instance, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Configurer) ApplyConfigFiles(instance string, files ...string) error {
	configs := make([]*cloudinit.Config, len(files))
	for i, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if !cloudinit.HasComment(data) {
			return fmt.Errorf("file %s does not start with %s", file, cloudinit.Comment)
		}
		var config cloudinit.Config
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		configs[i] = &config
	}
	for i, config := range configs {
		err := t.Apply(instance, config)
		if err != nil {
			return fmt.Errorf("%s: %w", files[i], err)
		}
	}
	return nil
}
