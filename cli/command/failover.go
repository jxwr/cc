package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/utils"
)

var FailoverCommand = cli.Command{
	Name:   "failover",
	Usage:  "failover <id>",
	Action: failoverAction,
}

func failoverAction(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	addr := context.GetLeaderAddr()
	extraHeader := &utils.ExtraHeader{
		User:  context.Config.User,
		Role:  context.Config.Role,
		Token: context.Config.Token,
	}

	url := "http://" + addr + api.NodeSetAsMasterPath
	nodeid, err := context.GetId(c.Args()[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	req := api.FailoverTakeoverParams{
		NodeId: nodeid,
	}
	resp, err := utils.HttpPostExtra(url, req, 5*time.Second, extraHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
