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

	extraHeader := &utils.ExtraHeader{
		User:  context.Config.User,
		Role:  context.Config.Role,
		Token: context.Config.Token,
	}

	url := "http://" + addr + api.RebalancePath

	req := api.RebalanceParams{
		Method:       "default",
		ShowPlanOnly: false,
	}
	resp, err := utils.HttpPostExtra(url, req, 5*time.Second, extraHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
