package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/utils"
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

	extraHeader := &utils.ExtraHeader{
		User:  context.Config.User,
		Role:  context.Config.Role,
		Token: context.Config.Token,
	}

	url := "http://" + addr + api.NodeReplicatePath
	cnodeid, err := context.GetId(c.Args()[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	pnodeid, err := context.GetId(c.Args()[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	req := api.ReplicateParams{
		ChildId:  cnodeid,
		ParentId: pnodeid,
	}
	resp, err := utils.HttpPostExtra(url, req, 5*time.Second, extraHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
