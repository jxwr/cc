package command

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
)

var AppInfoCommand = cli.Command{
	Name:   "appinfo",
	Usage:  "appinfo",
	Action: appInfoAction,
}

func appInfoAction(c *cli.Context) {
	if len(c.Args()) != 0 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	fmt.Println(context.GetAppInfo())
}
