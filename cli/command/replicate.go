package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
	"time"
)

var ReplicateCommand = cli.Command{
	Name:   "replicate",
	Usage:  "replicate <childid> <parentid>",
	Action: replicateAction,
}

func replicateAction(c *cli.Context) {
	if len(c.Args()) != 2 {
		fmt.Println("Error Usage")
		return
	}
	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.NodeReplicatePath
	cnodeid := c.Args()[0]
	pnodeid := c.Args()[1]

	req := api.ReplicateParams{
		ChildId:  cnodeid,
		ParentId: pnodeid,
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
