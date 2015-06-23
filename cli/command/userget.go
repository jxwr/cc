package command

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
)

var UserGetCommand = cli.Command{
	Name:   "userget",
	Usage:  "userget",
	Action: userGetAction,
	Flags: []cli.Flag{
		cli.StringFlag{"u,username", "", "username"},
	},
	Description: `
    get user token from zookeeper
    `,
}

func userGetAction(c *cli.Context) {
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

	if username == "" {
		fmt.Println("-u,username must be assigned")
		return
	}

	token, _, err := context.GetUser(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("User:%s\nToken:%s\n", username, token)
}
