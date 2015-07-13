package command

import (
	"encoding/json"
	"time"

	"github.com/codegangsta/cli"

	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/controller/command"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/meta"
	"github.com/ksarch-saas/cc/utils"
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

func FailoverRecords() ([]*meta.FailoverRecord, error) {
	zconn, _, err := meta.DialZk(context.ZkAddr)
	children, stat, err := zconn.Children("/r3/failover/history")
	if err != nil {
		return nil, err
	}
	if stat.NumChildren == 0 {
		return nil, nil
	}

	var records []*meta.FailoverRecord
	for _, file := range children {
		data, _, err := zconn.Get("/r3/failover/history/" + file)
		if err != nil {
			return nil, err
		}
		var record meta.FailoverRecord
		err = json.Unmarshal([]byte(data), &record)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}
	return records, nil
}

func showFailoverHistory() {
	records, err := FailoverRecords()
	if err != nil {
		Put(err)
		return
	}
	for _, r := range records {
		Putf("%s %s %s %6s %17s %19s %s\n",
			r.Timestamp.Format("2006/01/02 15:04:05"),
			r.AppName, r.NodeId, r.Role, r.Tag, r.NodeAddr, r.Ranges)
	}
}

func printShowUsage() {
	Put("List of show subcommands:\n")
	Put("show nodes    -- Show the nodes info group by replicaset")
	Put("show slots    -- Show the ranges of master nodes")
	Put("show tasks    -- Show migrating tasks")
	Put("show failover -- Show failover history records")
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
	case "failover":
		showFailoverHistory()
	default:
		printShowUsage()
	}
}
