package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var TakeoverCommand = cli.Command{
	Name:   "takeover",
	Usage:  "takeover <id>",
	Action: takeoverAction,
}

func takeoverAction(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.FailoverTakeoverPath
	nodeid, err := context.GetId(c.Args()[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	req := api.FailoverTakeoverParams{
		NodeId: nodeid,
	}
	resp, err := utils.HttpPost(url, req, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
