package command

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
)

var AppListCommand = cli.Command{
	Name:   "applist",
	Usage:  "applist",
	Action: appListAction,
	Description: `
    list all apps
    `,
}

func appListAction(c *cli.Context) {
	if len(c.Args()) != 0 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	apps, err := context.ListApp()
	if err != nil {
		fmt.Println(err)
		return
	}
	for idx, app := range apps {
		fmt.Printf("%d:\t%s\n", idx+1, app)
	}
	fmt.Printf("Total: %d app(s)\n", len(apps))
}
