package migrate

import (
	"fmt"

	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/topo"
)

var rebalanceTask *RebalanceTask

type MigratePlan struct {
	SourceId string       `json:"source_id"`
	TargetId string       `json:"target_id"`
	Ranges   []topo.Range `json:"ranges"`
	task     *MigrateTask
}

type RebalanceTask struct {
	Plans []*MigratePlan `json:"plans"`
}

type Rebalancer func(ss []*topo.Node, ts []*topo.Node) []*MigratePlan

var RebalancerTable = map[string]Rebalancer{
	"default": CutTailRebalancer,
	"cuttail": CutTailRebalancer,
}

func GenerateRebalancePlan(method string, cluster *topo.Cluster, targetIds []string) ([]*MigratePlan, error) {
	rss := cluster.ReplicaSets()
	regions := meta.AllRegions()

	ss := []*topo.Node{}          // 有slots的Master
	tm := map[string]*topo.Node{} // 空slots的Master

	for _, rs := range rss {
		master := rs.Master()
		// 忽略主挂掉和region覆盖不全的rs
		if master.Fail || !rs.CoverAllRegions(regions) {
			continue
		}
		if master.Empty() {
			tm[master.Id] = master
		} else {
			ss = append(ss, master)
		}
	}

	var ts []*topo.Node
	if len(targetIds) == 0 {
		for _, node := range tm {
			ts = append(ts, node)
		}
	} else {
		for _, id := range targetIds {
			if tm[id] == nil {
				return nil, fmt.Errorf("Master %s not found.", id)
			}
			ts = append(ts, tm[id])
		}
	}

	rebalancer := RebalancerTable[method]
	if rebalancer == nil {
		return nil, fmt.Errorf("Rebalancing method %s not exist.", method)
	}
	plans := rebalancer(ss, ts)
	return plans, nil
}
