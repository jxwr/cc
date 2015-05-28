package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var ShowCommand = cli.Command{
	Name:   "show",
	Usage:  "show [tasks]",
	Action: showAction,
}

func showMigrationTasks() {
	addr := context.GetLeaderAddr()
	url := "http://" + addr + api.FetchMigrationTasksPath
	resp, err := utils.HttpGet(url, nil, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	var res command.FetchMigrationTasksResult
	err = utils.InterfaceToStruct(resp.Body, &res)

	if len(res.Plans) == 0 {
		fmt.Println("No migration tasks.")
		return
	}

	for _, plan := range res.Plans {
		fmt.Printf("[%s:%d] %s->%s %v\n",
			plan.State, plan.CurrSlot, plan.SourceId, plan.TargetId, plan.Ranges)
	}
}

func showAction(c *cli.Context) {
	args := c.Args()
	if len(args) == 0 {
		fmt.Println(ErrInvalidParameter)
	}
	cmd := args[0]
	switch cmd {
	case "tasks", "task":
		showMigrationTasks()
	}
}
