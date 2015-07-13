package command

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ksarch-saas/cc/cli/context"
)

var AppDelCommand = cli.Command{
	Name:   "appdel",
	Usage:  "appdel",
	Action: appDelAction,
	Description: `
    [Warning]
    delete the app from zookeeper`,
}

func appDelAction(c *cli.Context) {
	super, err := context.CheckSuperPerm(context.Config.User)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !super {
		fmt.Println("You have no permission to this operation")
		return
	}
	appname := context.GetAppName()

	fmt.Printf("Type %s to continue: ", "yes")
	var cmd string
	fmt.Scanf("%s\n", &cmd)
	if cmd != "yes" {
		return
	}
	_, version, err := context.GetApp(appname)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = context.DelApp(appname, version)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("Delete %s success\n", appname)
	}
}
