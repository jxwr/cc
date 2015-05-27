package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var ReplicateCommand = cli.Command{
	Name:   "replicate",
	Usage:  "replicate <childid> <parentid>",
	Action: replicateAction,
}

func replicateAction(c *cli.Context) {
	if len(c.Args()) != 2 {
		fmt.Println(ErrInvalidParameter)
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
	resp, err := utils.HttpPost(url, req, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
