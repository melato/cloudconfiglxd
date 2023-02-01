package main

import (
	_ "embed"
	"fmt"

	"melato.org/command"
	"melato.org/command/usage"
	"melato.org/lxdcloudinit"
)

//go:embed version
var version string

//go:embed usage.yaml
var usageData []byte

func main() {
	cmd := &command.SimpleCommand{}
	var app lxdcloudinit.App
	cmd.Command("apply").Flags(&app).RunFunc(app.Apply)
	cmd.Command("version").RunFunc(func() { fmt.Println(version) })

	usage.Apply(cmd, usageData)
	command.Main(cmd)
}
