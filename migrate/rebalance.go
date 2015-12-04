package migrate

import (
	"fmt"
	"time"

	"github.com/ksarch-saas/cc/meta"
	"github.com/ksarch-saas/cc/topo"
)

type RebalanceTask struct {
	Plans     []*MigratePlan
	StartTime *time.Time
	EndTime   *time.Time
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
		master := rs.Master
		// 忽略主挂掉和region覆盖不全的rs
		if master.Fail || !rs.IsCoverAllRegions(regions) || master.Free {
			continue
		}
		if master.Empty() {
			tm[master.Id] = master
		} else {
			ss = append(ss, master)
		}
	}

	var ts []*topo.Node
	// 如果没传TargetId，则选择所有可以作为迁移目标的rs
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

	if len(ts) == 0 {
		return nil, fmt.Errorf("No available empty target replicasets.")
	}

	rebalancer := RebalancerTable[method]
	if rebalancer == nil {
		return nil, fmt.Errorf("Rebalancing method %s not exist.", method)
	}
	plans := rebalancer(ss, ts)

	return plans, nil
}
