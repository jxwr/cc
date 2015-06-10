package command

import (
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

const taskCommandUsage = "task <pause|resume|cancel> <sourceId>"

// Task actions
var TaskCommand = cli.Command{
	Name:   "task",
	Usage:  taskCommandUsage,
	Action: taskAction,
}

func doTaskAction(path, sourceId string) {
	addr := context.GetLeaderAddr()
	url := "http://" + addr + path
	nodeid, err := context.GetId(sourceId)
	if err != nil {
		Put(err)
		return
	}
	req := api.MigrateActionParams{
		SourceId: nodeid,
	}
	resp, err := utils.HttpPost(url, req, 5*time.Second)
	if err != nil {
		Put(err)
		return
	}
	ShowResponse(resp)
}

func taskAction(c *cli.Context) {
	if len(c.Args()) != 2 {
		Put(ErrInvalidParameter)
		return
	}

	action := c.Args()[0]
	sourceId := c.Args()[1]
	switch action {
	case "pause":
		doTaskAction(api.MigratePausePath, sourceId)
	case "resume":
		doTaskAction(api.MigrateResumePath, sourceId)
	case "cancel":
		doTaskAction(api.MigrateCancelPath, sourceId)
	default:
		Put(ErrInvalidParameter, "usage: \n"+taskCommandUsage)
	}
}
