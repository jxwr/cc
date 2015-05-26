package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
	"strings"
	"time"
)

var RebalanceCommand = cli.Command{
	Name:   "rebalance",
	Usage:  "rebalance <targetIds>",
	Action: rebalanceAction,
}

func rebalanceAction(c *cli.Context) {
	fmt.Println(c.Args())
	if len(c.Args()) != 1 {
		fmt.Println("Error Usage")
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
