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

	url := "http://" + addr + api.NodeMeetPath
	nodeid, err := context.GetId(c.Args()[0])
	if err != nil {
		Put(err)
		return
	}

	req := api.MeetNodeParams{
		NodeId: nodeid,
	}
	resp, err := utils.HttpPost(url, req, 5*time.Second)
	if err != nil {
		Put(err)
		return
	}
	ShowResponse(resp)
}
