package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
	"time"
)

var ChmodCommand = cli.Command{
	Name:   "chmod",
	Usage:  "change node read write state",
	Action: chmodAction,
	Flags: []cli.Flag{
		cli.BoolFlag{"r", "read state"},
		cli.BoolFlag{"w", "write state"},
	},
}

func chmodAction(c *cli.Context) {
	r := c.Bool("r")
	w := c.Bool("w")

	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.NodePermPath
	var act string
	var nodeid string
	var action string
	var perm string

	//-r -w
	if r || w {
		if len(c.Args()) != 1 || r == w {
			fmt.Println("Error Usage")
			return
		}
		action = "disable"
		nodeid = c.Args()[0]

		if r {
			perm = "read"
		} else {
			perm = "write"
		}
	} else {
		//+r +w
		if len(c.Args()) != 2 {
			fmt.Println("Error Usage")
			return
		}
		act = c.Args()[0]
		if string(act[0]) == "+" {
			action = "enable"
			nodeid = c.Args()[1]

			if string(act[1]) == "r" {
				perm = "read"
			} else if string(act[1]) == "w" {
				perm = "write"
			} else {
				fmt.Println("Error Usage")
				return
			}
		} else {
			fmt.Println("Error Usage")
			return
		}
	}

	req := api.ToggleModeParams{
		NodeId: nodeid,
		Action: action,
		Perm:   perm,
	}
	var resp api.MapResp
	fail, err := utils.HttpPost(url, req, &resp, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	if fail != nil {
		fmt.Println(fail)
		return
	}
	fmt.Println("OK")
}
