package command

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
)

var UserDelCommand = cli.Command{
	Name:   "userdel",
	Usage:  "userdel",
	Action: userDelAction,
	Flags: []cli.Flag{
		cli.StringFlag{"u,username", "", "username"},
	},
	Description: `
    delete user from zookeeper`,
}

func userDelAction(c *cli.Context) {
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

	fmt.Printf("Type %s to continue: ", "yes")
	var cmd string
	fmt.Scanf("%s\n", &cmd)
	if cmd != "yes" {
		return
	}
	_, version, err := context.GetUser(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = context.DelUser(username, version)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("Delete %s success\n", username)
	}
}
