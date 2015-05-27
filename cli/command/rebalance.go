package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var RebalanceCommand = cli.Command{
	Name:   "rebalance",
	Usage:  "rebalance <targetIds>",
	Action: rebalanceAction,
}

func rebalanceAction(c *cli.Context) {
	fmt.Println(c.Args())
	if len(c.Args()) != 1 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.RebalancePath
	nodes := strings.Fields(c.Args()[0])

	req := api.RebalanceParams{
		Method:       "default",
		TargetIds:    nodes,
		ShowPlanOnly: false,
	}
	resp, err := utils.HttpPost(url, req, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
