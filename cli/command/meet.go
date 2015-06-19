package command

import (
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var MeetCommand = cli.Command{
	Name:   "meet",
	Usage:  "meet <id>",
	Action: meetAction,
}

func meetAction(c *cli.Context) {
	if len(c.Args()) != 1 {
		Put(ErrInvalidParameter)
		return
	}
	addr := context.GetLeaderAddr()
	extraHeader := &utils.ExtraHeader{
		User:  context.Config.User,
		Role:  context.Config.Role,
		Token: context.Config.Token,
	}

	url := "http://" + addr + api.NodeMeetPath
	nodeid, err := context.GetId(c.Args()[0])
	if err != nil {
		Put(err)
		return
	}

	req := api.MeetNodeParams{
		NodeId: nodeid,
	}
	resp, err := utils.HttpPostExtra(url, req, 5*time.Second, extraHeader)
	if err != nil {
		Put(err)
		return
	}
	ShowResponse(resp)
}
