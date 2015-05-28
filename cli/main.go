package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/GeertJohan/go.linenoise"
	"github.com/codegangsta/cli"

	c "github.com/jxwr/cc/cli/command"
	"github.com/jxwr/cc/cli/command/initialize"
	"github.com/jxwr/cc/cli/context"
)

var cmds = []cli.Command{
	c.NodesCommand,
	c.ChmodCommand,
	c.FailoverCommand,
	c.TakeoverCommand,
	c.MigrateCommand,
	c.ReplicateCommand,
	c.RebalanceCommand,
	c.MeetCommand,
	c.ForgetAndResetCommand,
	c.AppInfoCommand,
}

var cmdmap = map[string]cli.Command{}

func init() {
	for _, cmd := range cmds {
		cmdmap[cmd.Name] = cmd
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		app := cli.NewApp()
		app.Name = "cli"
		app.Usage = "init a cluster"
		app.Commands = []cli.Command{initialize.Command}
		arg := append(os.Args)
		app.Run(arg)
		os.Exit(0)
	}
	if len(os.Args) == 1 {
		fmt.Println("Usage: cli <AppName> [<Command>] or cli init")
		os.Exit(1)
	}

	// Set context
	appName := os.Args[1]
	err := context.SetApp(appName, "127.0.0.1:2181")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// REPL
	if len(os.Args) == 2 {
		for {
			str, err := linenoise.Line(appName + "> ")
			if err != nil {
				if err == linenoise.KillSignalError {
					os.Exit(1)
				}
				fmt.Printf("Unexpected error: %s\n", err)
				os.Exit(1)
			}
			fields := strings.Fields(str)

			if len(fields) == 0 {
				continue
			}

			linenoise.AddHistory(str)

			cmd, ok := cmdmap[fields[0]]
			if !ok {
				fmt.Println("Error: unknown command.")
			}
			app := cli.NewApp()
			app.Name = cmd.Name
			app.Commands = []cli.Command{cmd}
			app.Run(append(os.Args[:1], fields...))
		}
	}

	// Command line
	if len(os.Args) > 2 {
		app := cli.NewApp()
		app.Name = "cli"
		app.Usage = "redis cluster cli"
		app.Commands = cmds
		app.Run(append(os.Args[:1], os.Args[2:]...))
	}
}
