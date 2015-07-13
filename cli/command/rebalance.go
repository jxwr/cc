package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/utils"
)

var RebalanceCommand = cli.Command{
	Name:   "rebalance",
	Usage:  "rebalance",
	Action: rebalanceAction,
}

func rebalanceAction(c *cli.Context) {
	if len(c.Args()) != 0 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.RebalancePath

	req := api.RebalanceParams{
		Method:       "default",
		ShowPlanOnly: false,
	}
	resp, err := utils.HttpPost(url, req, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
