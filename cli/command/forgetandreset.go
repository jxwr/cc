package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/utils"
)

var ForgetAndResetCommand = cli.Command{
	Name:   "forgetandreset",
	Usage:  "forgetandreset <id>",
	Action: forgetandresetAction,
}

func forgetandresetAction(c *cli.Context) {
	fmt.Println(c.Args())
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

	url := "http://" + addr + api.NodeForgetAndResetPath
	nodeid, err := context.GetId(c.Args()[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	req := api.ForgetAndResetNodeParams{
		NodeId: nodeid,
	}
	resp, err := utils.HttpPostExtra(url, req, 5*time.Second, extraHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
