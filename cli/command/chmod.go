package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var ChmodCommand = cli.Command{
	Name:   "chmod",
	Usage:  "chmod +r/-r/+w/-w <id>",
	Action: chmodAction,
	Flags: []cli.Flag{
		cli.BoolFlag{"r", "read state"},
		cli.BoolFlag{"w", "write state"},
	},
	Description: `
    change node read write perm
    `,
}

func chmodAction(c *cli.Context) {
	r := c.Bool("r")
	w := c.Bool("w")

	addr := context.GetLeaderAddr()

	extraHeader := &utils.ExtraHeader{
		User:  context.Config.User,
		Role:  context.Config.Role,
		Token: context.Config.Token,
	}

	url := "http://" + addr + api.NodePermPath
	var act string
	var nodeid string
	var action string
	var perm string
	var err error

	//-r -w
	if r || w {
		if len(c.Args()) != 1 || r == w {
			fmt.Println(ErrInvalidParameter)
			return
		}
		action = "disable"
		nodeid, err = context.GetId(c.Args()[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		if r {
			perm = "read"
		} else {
			perm = "write"
		}
	} else {
		//+r +w
		if len(c.Args()) != 2 {
			fmt.Println(ErrInvalidParameter)
			return
		}
		act = c.Args()[0]
		if string(act[0]) == "+" {
			action = "enable"
			nodeid, err = context.GetId(c.Args()[1])
			if err != nil {
				fmt.Println(err)
				return
			}

			if string(act[1]) == "r" {
				perm = "read"
			} else if string(act[1]) == "w" {
				perm = "write"
			} else {
				fmt.Println(ErrInvalidParameter)
				return
			}
		} else {
			fmt.Println(ErrInvalidParameter)
			return
		}
	}

	req := api.ToggleModeParams{
		NodeId: nodeid,
		Action: action,
		Perm:   perm,
	}
	resp, err := utils.HttpPostExtra(url, req, 5*time.Second, extraHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
