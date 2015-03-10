package state

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/topo"
)

var (
	ErrNodeNotExist = errors.New("cluster: node not exist")
)

type ClusterState struct {
	version     int64                 // 更新消息处理次数
	replicaSets []*topo.ReplicaSet    // 分片
	nodeStates  map[string]*NodeState // 节点状态机
}

func NewClusterState() *ClusterState {
	cs := &ClusterState{
		version:     0,
		nodeStates:  map[string]*NodeState{},
		replicaSets: []*topo.ReplicaSet{},
	}
	return cs
}

func (cs *ClusterState) AllNodeStates() map[string]*NodeState {
	return cs.nodeStates
}

func (cs *ClusterState) UpdateRegionNodes(region string, nodes []*topo.Node) {
	cs.version++
	now := time.Now()

	log.Println("Update region", region, len(nodes), "nodes")

	// 添加不存在的节点，版本号+1
	for _, n := range nodes {
		if n.Region != region {
			continue
		}
		nodeState := cs.nodeStates[n.Id]
		if nodeState == nil {
			nodeState = NewNodeState(n, cs.version)
			cs.nodeStates[n.Id] = nodeState
		} else {
			nodeState.version = cs.version
			nodeState.node = n
		}
		nodeState.updateTime = now
	}

	// 删除已经下线的节点
	for id, n := range cs.nodeStates {
		if n.node.Region != region {
			continue
		}
		nodeState := cs.nodeStates[id]
		if nodeState.version != cs.version {
			delete(cs.nodeStates, id)
		}
	}

	// NB: 每次都干这个很低效，先这么着，看看效果
	cs.BuildReplicaSets()
}

// 重建ReplicaSet
func (cs *ClusterState) BuildReplicaSets() {
	replicaSets := []*topo.ReplicaSet{}

	// 先找出Master创建RS
	for _, ns := range cs.nodeStates {
		node := ns.node
		// 已经处理过的挂掉的Master
		if node.Fail && len(node.Ranges) == 0 {
			continue
		}
		if node.IsMaster() {
			rs := topo.NewReplicaSet()
			rs.SetMaster(node)
			replicaSets = append(replicaSets, rs)
		}
	}

	// 把Slave塞进去
	for _, ns := range cs.nodeStates {
		node := ns.node
		if !node.IsMaster() {
			master := cs.FindNode(node.ParentId)
			// 出现这种情况，可能是有的地域没有汇报拓扑信息
			// 初始化时或节点变更时可能出现
			if master == nil {
				fmt.Printf("parent not exist failed: %s %s\n", node.ParentId, node.Addr())
				return
			}

			for _, rs := range replicaSets {
				if rs.Master() == master {
					rs.AddSlave(node)
				}
			}
		}
	}

	cs.replicaSets = replicaSets
}

func (cs *ClusterState) FindNode(nodeId string) *topo.Node {
	ns := cs.FindNodeState(nodeId)
	if ns == nil {
		return nil
	}
	return ns.node
}

func (cs *ClusterState) FindNodeState(nodeId string) *NodeState {
	return cs.nodeStates[nodeId]
}

func (cs *ClusterState) DebugDump() {
	var keys []string
	for k := range cs.nodeStates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("Cluster Debug Information:")
	for _, key := range keys {
		fmt.Print("  ")
		cs.nodeStates[key].DebugDump()
	}
}

func (cs *ClusterState) FindReplicaSetByNode(nodeId string) *topo.ReplicaSet {
	for _, rs := range cs.replicaSets {
		if rs.HasNode(nodeId) {
			return rs
		}
	}
	return nil
}

/// helpers

// 获取分片内ReplOffset最大的节点ID
func (cs *ClusterState) MaxReploffSlibing(nodeId string, slaveOnly bool) (string, error) {
	rs := cs.FindReplicaSetByNode(nodeId)
	if rs == nil {
		return "", ErrNodeNotExist
	}

	rmap := cs.FetchReplOffsetInReplicaSet(rs)

	var maxVal int64 = -1
	maxId := ""
	for id, val := range rmap {
		node := cs.FindNode(id)
		if slaveOnly && node.IsMaster() {
			continue
		}
		if val > maxVal {
			maxVal = val
			maxId = id
		}
	}

	return maxId, nil
}

type reploff struct {
	NodeId string
	Offset int64
}

// 失败返回-1
func fetchReplOffset(addr string) int64 {
	info, err := redis.FetchInfo(addr, "Replication")
	if err != nil {
		return -1
	}
	if info.Get("role") == "master" {
		offset, err := info.GetInt64("master_repl_offset")
		if err != nil {
			return -1
		} else {
			return offset
		}
	}
	offset, err := info.GetInt64("slave_repl_offset")
	if err != nil {
		return -1
	}
	return offset
}

// 获取分片内ReplOffset节点，包括Master
func (cs *ClusterState) FetchReplOffsetInReplicaSet(rs *topo.ReplicaSet) map[string]int64 {
	nodes := rs.AllNodes()
	c := make(chan reploff, len(nodes))

	for _, node := range nodes {
		go func(id, addr string) {
			offset := fetchReplOffset(addr)
			c <- reploff{id, offset}
		}(node.Id, node.Addr())
	}

	rmap := map[string]int64{}
	for i := 0; i < len(nodes); i++ {
		off := <-c
		rmap[off.NodeId] = off.Offset
	}
	return rmap
}

func (cs *ClusterState) RunFailoverTask(oldMasterId, newMasterId string) {
	new := cs.FindNodeState(newMasterId)
	old := cs.FindNodeState(oldMasterId)

	// 通过新主广播消息
	redis.DisableRead(new.Addr(), old.Id())
	redis.DisableWrite(new.Addr(), old.Id())

	c := make(chan error, 1)
	go func() {
		c <- redis.SetAsMasterWaitSyncDone(new.Addr(), true)
	}()

	select {
	case err := <-c:
		if err != nil {
			log.Printf("Failover finished with error(%v)\n", err)
		} else {
			log.Printf("Failover success\n", err)
		}
		old.AdvanceFSM(cs, CMD_FAILOVER_END_SIGNAL)
	case <-time.After(20 * time.Minute):
		log.Printf("Failover finished with error(timedout)\n")
		old.AdvanceFSM(cs, CMD_FAILOVER_END_SIGNAL)
	}

	// 打开新主的写入，因为给slave加Write没有效果
	// 所以即便Failover失败，也不会产生错误
	redis.DisableWrite(new.Addr(), new.Id())
}
