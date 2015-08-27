package initialize

import (
	"fmt"
	"github.com/ksarch-saas/cc/redis"
	"strconv"
	"time"
)

func isAlive(node *Node) bool {
	addr := fmt.Sprintf("%s:%s", node.Ip, node.Port)
	return redis.IsAlive(addr)
}

func isEmpty(node *Node) bool {
	return true
}

func isMaster(node *Node) bool {
	addr := fmt.Sprintf("%s:%s", node.Ip, node.Port)
	ri, _ := redis.FetchInfo(addr, "Replication")
	role := ri.Get("role")
	return role == "master"
}

func resetNodes(nodes []*Node) (string, error) {
	var ChanErr chan error = make(chan error)
	var ret_err error = nil
	for _, node := range nodes {
		inner := func(node *Node) {
			var err error
			addr := fmt.Sprintf("%s:%s", node.Ip, node.Port)
			if isMaster(node) {
				_, err = redis.FlushAll(addr)
			}
			_, err = redis.ClusterReset(addr, false)
			ChanErr <- err
		}
		go inner(node)
	}
	for i := 0; i < len(nodes); i++ {
		err := <-ChanErr
		if err != nil {
			ret_err = err
		}
	}
	return "", ret_err
}

func clusterNodes(node *Node) (string, error) {
	addr := fmt.Sprintf("%s:%s", node.Ip, node.Port)
	resp, err := redis.ClusterNodesWithoutExtra(addr)
	return resp, err
}

func meetEach(nodes []*Node) {
	for _, n1 := range nodes {
		for _, n2 := range nodes {
			if n1 != n2 {
				addr := fmt.Sprintf("%s:%s", n1.Ip, n1.Port)
				newPort, _ := strconv.Atoi(n2.Port)
				redis.ClusterMeet(addr, n2.Ip, newPort)
			}
		}
	}
}

func addSlotRange(node *Node) (string, error) {
	addr := fmt.Sprintf("%s:%s", node.Ip, node.Port)
	var start int
	var end int
	fmt.Sscanf(node.SlotsRange, "%d-%d", &start, &end)
	return redis.AddSlotRange(addr, start, end)
}

func setReplicas(slaves []*Node) (string, error) {
	var resp string
	var err error
	for _, slave := range slaves {
		addr := fmt.Sprintf("%s:%s", slave.Ip, slave.Port)
		resp, err = redis.ClusterReplicate(addr, slave.MasterId)
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func checkClusterInfo(nodes []*Node) bool {
	retry := 3
	inner := func(nodes []*Node) bool {
		var (
			clusterstate           string
			cluster_slots_assigned int
			cluster_slots_ok       int
			cluster_slots_pfail    int
			cluster_slots_fail     int
			cluster_known_nodes    int
			cluster_size           int
		)

		for idx, node := range nodes {
			addr := fmt.Sprintf("%s:%s", node.Ip, node.Port)
			ci, err := redis.FetchClusterInfo(addr)
			if err != nil {
				return false
			}
			if idx == 0 {
				clusterstate = ci.ClusterState
				cluster_slots_assigned = ci.ClusterSlotsAssigned
				cluster_slots_ok = ci.ClusterSlotsOk
				if cluster_slots_ok != 16384 {
					return false
				}
				cluster_slots_pfail = ci.ClusterSlotsPfail
				cluster_slots_fail = ci.ClusterSlotsFail
				cluster_known_nodes = ci.ClusterKnownNodes
				cluster_size = ci.ClusterSize
			} else {
				if clusterstate != ci.ClusterState ||
					cluster_slots_assigned != ci.ClusterSlotsAssigned ||
					cluster_slots_ok != ci.ClusterSlotsOk ||
					cluster_slots_pfail != ci.ClusterSlotsPfail ||
					cluster_slots_fail != ci.ClusterSlotsFail ||
					cluster_known_nodes != ci.ClusterKnownNodes ||
					cluster_size != ci.ClusterSize {
					return false
				}
			}
		}
		return true
	}

	for retry > 0 {
		time.Sleep(time.Second * 5)
		fmt.Printf("checking %d times\n", 4-retry)
		if inner(nodes) {
			return true
		}
		retry = retry - 1
	}
	return false
}

func rwMasterState(node *Node) (string, error) {
	addr := fmt.Sprintf("%s:%s", node.Ip, node.Port)
	resp, err := redis.EnableRead(addr, node.Id)
	if err != nil {
		return resp, err
	}
	resp, err = redis.EnableWrite(addr, node.Id)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func rwSlaveState(node *Node) (string, error) {
	addr := fmt.Sprintf("%s:%s", node.Ip, node.Port)
	resp, err := redis.EnableRead(addr, node.Id)
	if err != nil {
		return resp, err
	}
	resp, err = redis.DisableWrite(addr, node.Id)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func rwReplicasState(nodes []*Node) error {
	for _, node := range nodes {
		_, err := rwSlaveState(node)
		if err != nil {
			return err
		}
	}
	return nil
}
