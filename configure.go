package lxdcloudinit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"melato.org/cloudinit"
	"melato.org/lxdcloudinit/lxdutil"
	"melato.org/yaml"
)

var Trace bool

type Configurer struct {
	Server      lxd.InstanceServer
	OS          cloudinit.OS
	Log         io.Writer
	instance    string
	createdDirs map[string]struct{}
}

// NewConfigurer creates a Configurer for an instance.
// The Configurer should not be reused for other instances.
func NewConfigurer(server lxd.InstanceServer, instance string) *Configurer {
	t := &Configurer{Server: server, instance: instance}
	t.createdDirs = make(map[string]struct{})
	return t
}

func (t *Configurer) log(format string, args ...any) {
	if t.Log != nil {
		fmt.Fprintf(t.Log, format, args...)
	}
}

func (t *Configurer) ensureDirExists(dir string) error {
	if dir == "/" || dir == "." {
		return nil
	}
	_, exists := t.createdDirs[dir]
	if exists {
		return nil
	}
	err := t.exec("", "mkdir", "-p", dir)
	if err != nil {
		return err
	}
	for d := dir; !(d == "." || d == "/"); d = filepath.Dir(d) {
		t.createdDirs[d] = struct{}{}
	}
	return nil
}

func (t *Configurer) WriteFile(f *cloudinit.File) error {
	var args lxd.InstanceFileArgs
	if f.Permissions != "" {
		mode, err := strconv.ParseInt(f.Permissions, 8, 32)
		if err != nil {
			return err
		}
		args.Mode = int(mode)
	} else {
		// does cloud-init specify default permissions?
		args.Mode = 0644
	}
	t.log("write file: %s\n", f.Path)
	dir := filepath.Dir(f.Path)
	err := t.ensureDirExists(dir)
	if err != nil {
		return err
	}
	args.Content = strings.NewReader(f.Content)
	err = t.Server.CreateInstanceFile(t.instance, f.Path, args)
	if err != nil {
		return lxdutil.AnnotateLXDError(f.Path, err)
	}
	if f.Owner != "" {
		err := t.exec("", "chown", f.Owner, f.Path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Configurer) Apply(config *cloudinit.Config) error {
	err := t.InstallPackages(config.Packages)
	if err != nil {
		return err
	}
	for _, f := range config.Files {
		err := t.WriteFile(f)
		if err != nil {
			return err
		}
	}
	err = t.RunCommands(config.Runcmd)
	if err != nil {
		return err
	}
	return nil
}

func (t *Configurer) ApplyConfigFiles(files ...string) error {
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
		err := t.Apply(config)
		if err != nil {
			return fmt.Errorf("%s: %w", files[i], err)
		}
	}
	return nil
}

func (t *Configurer) RunCommands(commands []cloudinit.Command) error {
	for _, command := range commands {
		err := t.runCommand(command)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Configurer) runCommand(command cloudinit.Command) error {
	script, isScript := cloudinit.CommandScript(command)
	if isScript {
		return t.exec(script, "/bin/sh")
	}
	args, isArgs := cloudinit.CommandArgs(command)
	if isArgs {
		return t.exec("", args...)
	}
	return fmt.Errorf("invalid command type: %T", command)
}

// NopWriteCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

func (t *Configurer) exec(input string, execArgs ...string) error {
	if len(execArgs) == 0 {
		return fmt.Errorf("empty command")
	}
	if t.Log != nil {
		var suffix string
		if input != "" {
			suffix = " << ---"
		}
		fmt.Printf("%s%s\n", strings.Join(execArgs, " "), suffix)
		if input != "" {
			fmt.Printf("%s\n---\n", input)
		}
	}
	var post api.InstanceExecPost
	post.Command = execArgs
	post.WaitForWS = true

	var args lxd.InstanceExecArgs
	if t.Log != nil {
		args.Stderr = NopWriteCloser(t.Log)
	} else {
		args.Stderr = NopWriteCloser(os.Stderr)
	}
	if t.Log != nil {
		args.Stdout = NopWriteCloser(t.Log)
	}

	if input != "" {
		args.Stdin = io.NopCloser(strings.NewReader(input))
	}
	op, err := t.Server.ExecInstance(t.instance, post, &args)
	if err != nil {
		return lxdutil.AnnotateLXDError(t.instance, err)
	}
	err = op.Wait()
	if err != nil {
		return lxdutil.AnnotateLXDError(t.instance, err)
	}
	return nil
}

func (t *Configurer) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}
	if t.OS == nil {
		return requireOSError("cannot install packages")
	}
	commands := make([]cloudinit.Command, 0, len(packages))
	for _, pkg := range packages {
		commands = append(commands, t.OS.InstallPackageCommand(pkg))
	}
	return t.RunCommands(commands)
}

func requireOSError(msg string) error {
	return fmt.Errorf("%s.  Missing OS", msg)
}
