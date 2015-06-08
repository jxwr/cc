package initialize

import (
	"errors"
	"fmt"
	"github.com/codeskyblue/go-sh"
	"strings"
)

var (
	ERR_NODES_NUMBER = errors.New("Nodes number invalid")

	REDIS_CLUSTER_SLOTS = 16384
)

type Node struct {
	Id         string
	Ip         string
	Port       string
	Tag        string
	LogicMR    string
	Role       string
	MasterId   string
	SlotsRange string
	Alive      bool
	Empty      bool
	Met        bool
	Chosen     bool
}

func splitLineFunc(r rune) bool {
	return r == '\r' || r == '\n'
}

func SplitLine(str string) []string {
	return strings.FieldsFunc(str, splitLineFunc)
}

/* get nodes from osp service */
func getNodes(service string) ([]*Node, error) {
	res, err := sh.Command("get_instance_by_service", "-i", "-p", service).Output()
	//res, err := sh.Command("cat", service).Output()
	str := string(res)
	str = strings.TrimSpace(str)
	/*hostname ip port*/
	lines := SplitLine(str)
	nodes := []*Node{}
	for _, line := range lines {
		xs := strings.Fields(line)
		node := Node{
			Ip:   xs[1],
			Port: xs[2],
		}
		nodes = append(nodes, &node)
	}
	return nodes, err
}

func validateProcess(nodes []*Node) bool {
	res := true
	for _, node := range nodes {
		if node.Met {
			fmt.Printf("%s:%s has jioned a cluster\n", node.Ip, node.Port)
			res = false
		} else if node.Alive == false {
			fmt.Printf("%s:%s is not alive\n", node.Ip, node.Port)
			res = false
		} else if node.Id == "" {
			fmt.Printf("%s:%s in wrong state\n", node.Ip, node.Port)
			res = false
		} else if node.Role != "master" {
			fmt.Printf("%s:%s is not a master\n", node.Ip, node.Port)
			res = false
		}
	}
	return res
}

func checkAndSetState(node *Node) {
	if node.Alive == false {
		return
	}
	info, err := clusterNodes(node)
	if err != nil {
		return
	}
	if len(SplitLine(info)) > 1 {
		node.Met = true
		return
	}

	//set info state
	cols := strings.Fields(info)
	if len(cols) != 8 {
		return
	}
	node.Id = cols[0]
	role := cols[2]
	node.Role = strings.Split(role, ",")[1]
}

//try my best to choose num node in differrnt host
func chooseMaster(nodes []*Node, logicRoom string, num int) []*Node {
	var res []*Node
	ipSet := make(map[string]bool)
	for _, node := range nodes {
		if num == 0 {
			break
		}

		if node.Chosen || node.LogicMR != logicRoom {
			continue
		}

		if ipSet[node.Ip] {
			continue
		}

		res = append(res, node)
		ipSet[node.Ip] = true
		node.Chosen = true

		num = num - 1
	}

	for _, node := range nodes {
		if num == 0 {
			break
		}
		if node.Chosen || node.LogicMR != logicRoom {
			continue
		}
		res = append(res, node)
		node.Chosen = true

		num = num - 1
	}
	return res
}

func getFreeNodes(nodes []*Node, logicRoom string) []*Node {
	var res []*Node
	for _, node := range nodes {
		if node.Chosen || node.LogicMR != logicRoom {
			continue
		}
		res = append(res, node)
	}
	return res
}

//try my best to choose num nodes in differrnt hosts
func getAndRemoveReplicas(nodes []*Node, num int, master *Node) ([]*Node, []*Node) {
	if num > len(nodes) {
		return nil, nodes
	}
	ipSet := make(map[string]bool)
	var choose []*Node
	var left []*Node

	for _, node := range nodes {
		if num == 0 {
			break
		}
		if node.Chosen || node.Ip == master.Ip {
			continue
		}
		if node.LogicMR == master.LogicMR {
			continue
		}
		if ipSet[node.Ip] {
			continue
		}
		choose = append(choose, node)
		node.Chosen = true
		ipSet[node.Ip] = true
		num = num - 1
	}

	for _, node := range nodes {
		if num == 0 {
			break
		}
		if node.Chosen {
			continue
		}
		choose = append(choose, node)
		node.Chosen = true
		num = num - 1
	}

	for _, node := range nodes {
		for _, cnode := range choose {
			if node.Id == cnode.Id {
				continue
			}
			left = append(left, node)
		}
	}
	return choose, left
}

func buildCluster(nodes []*Node, replicas int, masterRooms, allRooms []string) ([]*Node, error) {
	masterNum := len(nodes) / (replicas + 1)

	//根据所有机房和主地域机房，统计非主地域机房
	var otherRooms []string
	for _, ar := range allRooms {
		isMasterRoom := false
		for _, mr := range masterRooms {
			if mr == ar {
				isMasterRoom = true
			}
		}
		if isMasterRoom == false {
			otherRooms = append(otherRooms, ar)
		}
	}

	//按master数目和逻辑机房数目平均分配每个逻辑机房的master数目
	var masterNodes []*Node
	avg := masterNum / len(masterRooms)
	for _, r := range masterRooms {
		//master尽量选择不同机器
		mNodes := chooseMaster(nodes, r, avg)
		masterNodes = append(masterNodes, mNodes...)
	}

	left := masterNum - avg*len(masterRooms)
	if left != 0 {
		mNodes := chooseMaster(nodes, masterRooms[0], left)
		masterNodes = append(masterNodes, mNodes...)
	}

	//统计每个机房剩余的node，用于分配salve
	repliNodes := map[string][]*Node{}
	for _, r := range allRooms {
		repliNodes[r] = getFreeNodes(nodes, r)
	}

	//合并主逻辑地域node
	var mRegionRepli []*Node
	for _, r := range masterRooms {
		mRegionRepli = append(mRegionRepli, repliNodes[r]...)
	}
	if len(mRegionRepli) > 0 && len(mRegionRepli)%masterNum != 0 {
		return nil, ERR_NODES_NUMBER
	}

	//非主地域node个数检查
	for _, r := range otherRooms {
		if len(repliNodes[r]) > 0 && len(repliNodes[r])%masterNum != 0 {
			return nil, ERR_NODES_NUMBER
		}
	}

	var rNodes []*Node
	//主地域
	avgRepli := len(mRegionRepli) / masterNum
	for _, master := range masterNodes {
		rNodes, mRegionRepli = getAndRemoveReplicas(mRegionRepli, avgRepli, master)
		//set masterid
		for _, node := range rNodes {
			node.MasterId = master.Id
			node.Role = "slave"
		}
	}
	//非主地域
	for _, r := range otherRooms {
		avgRepli = len(repliNodes[r]) / masterNum
		for _, master := range masterNodes {
			rNodes, repliNodes[r] = getAndRemoveReplicas(repliNodes[r], avgRepli, master)
			//set masterid
			for _, node := range rNodes {
				node.MasterId = master.Id
				node.Role = "slave"
			}
		}
	}

	return masterNodes, nil
}

func assignSlots(nodes []*Node) error {
	if len(nodes) == 0 {
		return ERR_NODES_NUMBER
	}
	step := REDIS_CLUSTER_SLOTS / len(nodes)
	left := REDIS_CLUSTER_SLOTS % len(nodes)
	start := 0
	end := 0

	for _, node := range nodes {
		end = start + step - 1
		if left > 0 {
			left = left - 1
			end = end + 1
		}
		node.SlotsRange = fmt.Sprintf("%d-%d", start, end)
		start = end + 1
	}
	return nil
}

func getSlaves(nodes []*Node, master *Node) []*Node {
	var res []*Node
	for _, node := range nodes {
		if node.MasterId == master.Id {
			res = append(res, node)
		}
	}
	return res
}
