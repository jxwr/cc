package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/topo"
	"github.com/jxwr/cc/utils"
)

var NodesCommand = cli.Command{
	Name:   "nodes",
	Usage:  "nodes",
	Action: nodesAction,
}

func toInterfaceSlice(nodes []*topo.Node) []interface{} {
	var interfaceSlice []interface{} = make([]interface{}, len(nodes))
	for i, node := range nodes {
		interfaceSlice[i] = node
	}
	return interfaceSlice
}

func nodesAction(c *cli.Context) {
	addr := context.GetLeaderAddr()
	url := "http://" + addr + api.FetchReplicaSetsPath

	resp, err := utils.HttpGet(url, nil, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	var rss command.FetchReplicaSetsResult
	err = json.Unmarshal(resp.Body, &rss)
	if err != nil {
		fmt.Println("Parse resp error:", err)
		return
	}
	sort.Sort(topo.ByMasterId(rss.ReplicaSets))
	var allNodes []*topo.Node
	for _, rs := range rss.ReplicaSets {
		allNodes = append(allNodes, rs.Master)
		for _, node := range rs.Slaves {
			allNodes = append(allNodes, node)
		}
	}

	utils.PrintJsonArray("table",
		[]string{"Id", "Ip", "Port", "Tag", "Role", "Readable", "Writable", "Free"},
		toInterfaceSlice(allNodes))
}
