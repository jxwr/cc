package state

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jxwr/cc/log"
	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/topo"
)

var (
	ErrNodeNotExist = errors.New("cluster: node not exist")
)

type ClusterState struct {
	version    int64                 // 更新消息处理次数
	cluster    *topo.Cluster         // 集群拓扑快照
	nodeStates map[string]*NodeState // 节点状态机
}

func NewClusterState() *ClusterState {
	cs := &ClusterState{
		version:    0,
		nodeStates: map[string]*NodeState{},
	}
	return cs
}

func (cs *ClusterState) AllNodeStates() map[string]*NodeState {
	return cs.nodeStates
}

func (cs *ClusterState) UpdateRegionNodes(region string, nodes []*topo.Node) {
	cs.version++
	now := time.Now()

	log.Verbosef("CLUSTER", "Update region %s %d nodes", region, len(nodes))

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

	// NB：低效？
	cs.BuildClusterSnapshot()
}

func (cs *ClusterState) GetClusterSnapshot() *topo.Cluster {
	return cs.cluster
}

func (cs *ClusterState) BuildClusterSnapshot() {
	// __CC__没什么意义，跟Region区别开即可
	cluster := topo.NewCluster("__CC__")
	for _, ns := range cs.nodeStates {
		cluster.AddNode(ns.node)
	}
	err := cluster.BuildReplicaSets()
	// 出现这种情况，很可能是启动时节点还不全
	if err != nil {
		log.Info("CLUSTER", "Build cluster snapshot failed ", err)
		return
	}
	cs.cluster = cluster
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
	if cs.cluster != nil {
		return cs.cluster.FindReplicaSetByNode(nodeId)
	} else {
		return nil
	}
}

/// helpers

// 获取分片内主地域中ReplOffset最大的节点ID
func (cs *ClusterState) MaxReploffSlibing(nodeId string, region string, slaveOnly bool) (string, error) {
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
		if node.Region != region {
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
			log.Eventf(old.Addr(), "Failover finished with error(%v)", err)
		} else {
			log.Eventf(old.Addr(), "Failover success, new master %s(%s)", new.Addr(), new.Id())
		}
		old.AdvanceFSM(cs, CMD_FAILOVER_END_SIGNAL)
	case <-time.After(20 * time.Minute):
		log.Eventf(old.Addr(), "Failover timedout, new master %s(%s)", new.Addr(), new.Id())
		// 判断是否主从是否已经切换
		old.AdvanceFSM(cs, CMD_FAILOVER_END_SIGNAL)
	}

	// 重新读取一次，因为可能已经更新了
	roleChanged := false
	node := cs.FindNode(newMasterId)
	if node.IsMaster() {
		roleChanged = true
	} else {
		info, err := redis.FetchInfo(node.Addr(), "Replication")
		if err == nil && info.Get("role") == "master" {
			for { // 等Cluster刷新
				node := cs.FindNode(newMasterId)
				if !node.IsMaster() {
					time.Sleep(1 * time.Second)
				} else {
					roleChanged = true
				}
			}
		}
	}

	if roleChanged {
		log.Eventf(old.Addr(), "New master %s(%s) role change success", node.Id, node.Addr())
		// 处理迁移过程中的异常问题，将故障节点（旧主）的slots转移到新主上
		oldNode := cs.FindNode(oldMasterId)
		if oldNode.Fail && oldNode.IsMaster() && len(oldNode.Ranges) != 0 {
			for _, r := range oldNode.Ranges {
				log.Eventf(old.Addr(), "Fix migration task %d-%d", r.Left, r.Right)

				for i := r.Left; i <= r.Right; i++ {
					for _, ns := range cs.AllNodeStates() {
						if ns.node.IsMaster() {
							log.Warningf(old.Addr(), "setslot %d node %s", i, new.Id())

							redis.SetSlot(ns.Addr(), i, redis.SLOT_NODE, new.Id())
						}
					}
				}
			}
		} else {
			log.Info(old.Addr(), "Good, no slot need to be fix after failover.")
		}
	} else {
		log.Warningf(old.Addr(), "Failover failed, please check cluster state.")
	}

	// 打开新主的写入，因为给slave加Write没有效果
	// 所以即便Failover失败，也不会产生错误
	redis.EnableWrite(new.Addr(), new.Id())
}
