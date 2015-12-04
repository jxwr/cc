package command

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ksarch-saas/cc/cli/context"
)

var UserAddCommand = cli.Command{
	Name:   "useradd",
	Usage:  "useradd",
	Action: userAddAction,
	Flags: []cli.Flag{
		cli.StringFlag{"u,username", "", "username"},
		cli.StringFlag{"r,role", "admin", "role"},
	},
	Description: `
    add user token to zookeeper
    `,
}

func userAddAction(c *cli.Context) {
	super, err := context.CheckSuperPerm(context.Config.User)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !super {
		fmt.Println("You have no permission to this operation")
		return
	}

	username := c.String("u")
	role := c.String("r")

	if username == "" {
		fmt.Println("-u,username must be assigned")
		return
	}

	token, err := context.AddUser(username, role)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Add %s success\nToken:%s\n", username, token)
}
