package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var FailoverCommand = cli.Command{
	Name:   "failover",
	Usage:  "failover <id>",
	Action: failoverAction,
}

func failoverAction(c *cli.Context) {
	fmt.Println(c.Args())
	if len(c.Args()) != 1 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.NodeSetAsMasterPath
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
