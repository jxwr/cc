package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
	"time"
)

var FailoverCommand = cli.Command{
	Name:   "failover",
	Usage:  "failover <id>",
	Action: failoverAction,
}

func failoverAction(c *cli.Context) {
	fmt.Println(c.Args())
	if len(c.Args()) != 1 {
		fmt.Println("Error Usage")
		return
	}
	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.NodeSetAsMasterPath
	nodeid := c.Args()[0]

	req := api.FailoverTakeoverParams{
		NodeId: nodeid,
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
