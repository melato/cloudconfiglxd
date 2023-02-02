package lxdcloudinit

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"melato.org/lxdcloudinit/lxdutil"
)

type InstanceConfigurer struct {
	Server      lxd.InstanceServer
	Log         io.Writer
	instance    string
	createdDirs map[string]struct{}
}

// NewInstanceConfigurer creates a BaseConfigurer for an instance.
// The configurer should not be reused for other instances.
func NewInstanceConfigurer(server lxd.InstanceServer, instance string) *InstanceConfigurer {
	t := &InstanceConfigurer{Server: server, instance: instance}
	t.createdDirs = make(map[string]struct{})
	return t
}

func (t *InstanceConfigurer) RunScript(script string) error {
	return t.exec(script, "/bin/sh")

}

func (t *InstanceConfigurer) RunCommand(args ...string) error {
	return t.exec("", args...)

}

func (t *InstanceConfigurer) exec(input string, execArgs ...string) error {
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

func (t *InstanceConfigurer) ensureDirExists(dir string) error {
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

func (t *InstanceConfigurer) WriteFile(path string, data []byte, perm fs.FileMode) error {
	var args lxd.InstanceFileArgs
	dir := filepath.Dir(path)
	err := t.ensureDirExists(dir)
	if err != nil {
		return err
	}
	args.Content = bytes.NewReader(data)
	err = t.Server.CreateInstanceFile(t.instance, path, args)
	if err != nil {
		return lxdutil.AnnotateLXDError(path, err)
	}
	return nil
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
