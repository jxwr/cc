package command

import (
	"fmt"
	"sort"
	"time"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/topo"
	"github.com/jxwr/cc/utils"
)

/// Show Nodes

type RNode struct {
	State      string
	Id         string
	ParentId   string
	Role       string
	Addr       string
	Fail       string
	Mode       string
	Tag        string
	Keys       int64
	Repl       string
	Link       string
	QPS        int
	NetIn      string
	NetOut     string
	UsedMemory string
}

func toReadable(node *topo.Node, state string) *RNode {
	if node == nil {
		return nil
	}
	n := &RNode{
		State:    state,
		Id:       node.Id,
		ParentId: node.ParentId,
		Tag:      node.Tag,
		Role:     "S",
		Fail:     "OK",
		Mode:     "--",
		Addr:     fmt.Sprintf("%s:%d", node.Ip, node.Port),
		Keys:     node.SummaryInfo.Keys,
		Link:     node.SummaryInfo.MasterLinkStatus,
		QPS:      node.SummaryInfo.InstantaneousOpsPerSec,
	}

	if node.Role == "master" {
		n.Role = "M"
	}
	if node.Fail {
		n.Fail = "Fail"
	}
	if node.Readable && node.Writable {
		n.Mode = "rw"
	}
	if node.Readable && !node.Writable {
		n.Mode = "r-"
	}
	if !node.Readable && node.Writable {
		n.Mode = "-w"
	}
	if node.IsMaster() {
		n.Link = "up"
	}
	n.UsedMemory = fmt.Sprintf("%0.2fG", float64(node.SummaryInfo.UsedMemory)/1024.0/1024.0/1024.0)
	n.NetIn = fmt.Sprintf("%.2fKbps", node.SummaryInfo.InstantaneousInputKbps)
	n.NetOut = fmt.Sprintf("%.2fKbps", node.SummaryInfo.InstantaneousOutputKbps)
	n.Repl = fmt.Sprintf("%d", node.ReplOffset)
	return n
}

func nodesToInterfaceSlice(nodes []*topo.Node, stateMap map[string]string) []interface{} {
	var interfaceSlice []interface{} = make([]interface{}, len(nodes))
	for i, node := range nodes {
		state := ""
		if node != nil {
			var ok bool
			state, ok = stateMap[node.Id]
			if !ok {
				state = "UNKNOWN"
			}
		}
		interfaceSlice[i] = toReadable(node, state)
	}
	return interfaceSlice
}

func showNodes() {
	addr := context.GetLeaderAddr()
	url := "http://" + addr + api.FetchReplicaSetsPath

	resp, err := utils.HttpGet(url, nil, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	var rss command.FetchReplicaSetsResult
	err = utils.InterfaceToStruct(resp.Body, &rss)
	if err != nil {
		fmt.Println(err)
		return
	}
	sort.Sort(topo.ByMasterId(rss.ReplicaSets))

	var allNodes []*topo.Node
	for i, rs := range rss.ReplicaSets {
		allNodes = append(allNodes, rs.Master)
		for _, node := range rs.Slaves {
			allNodes = append(allNodes, node)
		}
		if i < len(rss.ReplicaSets)-1 {
			allNodes = append(allNodes, nil)
		}
	}
	utils.PrintJsonArray("table",
		[]string{"State", "Mode", "Fail", "Role", "Id", "Tag", "Addr", "QPS",
			"UsedMemory", "Link", "Repl", "Keys", "NetIn", "NetOut"},
		nodesToInterfaceSlice(allNodes, rss.NodeStates))
}

/// Show Slots

type SlotsRow struct {
	Id     string
	Total  int
	Ranges string
}

func rowsToInterfaceSlice(rows []*SlotsRow) []interface{} {
	var interfaceSlice []interface{} = make([]interface{}, len(rows))
	for i, row := range rows {
		interfaceSlice[i] = row
	}
	return interfaceSlice
}

func showSlots() {
	addr := context.GetLeaderAddr()
	url := "http://" + addr + api.FetchReplicaSetsPath

	resp, err := utils.HttpGet(url, nil, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	var rss command.FetchReplicaSetsResult
	err = utils.InterfaceToStruct(resp.Body, &rss)
	if err != nil {
		fmt.Println(err)
		return
	}
	sort.Sort(topo.ByMasterId(rss.ReplicaSets))

	var rows []*SlotsRow
	for _, rs := range rss.ReplicaSets {
		rows = append(rows, &SlotsRow{rs.Master.Id, rs.Master.NumSlots(), topo.Ranges(rs.Master.Ranges).String()})
	}
	utils.PrintJsonArray("table", []string{"Id", "Total", "Ranges"}, rowsToInterfaceSlice(rows))
}
