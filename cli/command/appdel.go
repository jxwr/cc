package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
)

var AppDelCommand = cli.Command{
	Name:   "appdel",
	Usage:  "appdel",
	Action: appDelAction,
	Flags: []cli.Flag{
		cli.StringFlag{"n,appname", "", "appname"},
	},
}

func appDelAction(c *cli.Context) {
	appname := c.String("n")

	if appname == "" {
		fmt.Println("-n,appname must be assigned")
		os.Exit(-1)
	}
	fmt.Printf("Type %s to continue: ", "yes")
	var cmd string
	fmt.Scanf("%s\n", &cmd)
	if cmd != "yes" {
		os.Exit(-1)
	}
	_, version, err := context.GetApp(appname)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	err = context.DelApp(appname, version)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	} else {
		fmt.Printf("Delete %s success\n", appname)
	}
}
