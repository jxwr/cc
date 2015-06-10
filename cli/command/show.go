package command

import (
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
)

var ShowCommand = cli.Command{
	Name:   "show",
	Usage:  "show <tasks|nodes|slots>",
	Action: showAction,
}

func showMigrationTasks() {
	addr := context.GetLeaderAddr()
	url := "http://" + addr + api.FetchMigrationTasksPath
	resp, err := utils.HttpGet(url, nil, 5*time.Second)
	if err != nil {
		Put(err)
		return
	}
	var res command.FetchMigrationTasksResult
	err = utils.InterfaceToStruct(resp.Body, &res)

	if len(res.Plans) == 0 {
		Put("No migration tasks.")
		return
	}

	for _, plan := range res.Plans {
		Putf("[%s:%d] %s->%s %v\n",
			plan.State, plan.CurrSlot, plan.SourceId, plan.TargetId, plan.Ranges)
	}
}

func printShowUsage() {
	Put("List of show subcommands:\n")
	Put("show nodes -- Show the nodes info group by replicaset")
	Put("show slots -- Show the ranges of master nodes")
	Put("show tasks -- Show migrating tasks")
	Put()
}

func showAction(c *cli.Context) {
	args := c.Args()
	if len(args) == 0 {
		printShowUsage()
		return
	}
	cmd := args[0]
	switch cmd {
	case "tasks", "task":
		showMigrationTasks()
	case "nodes":
		showNodes()
	case "slots":
		showSlots()
	default:
		printShowUsage()
	}
}
