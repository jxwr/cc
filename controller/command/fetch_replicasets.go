package command

import (
	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/topo"
)

type FetchReplicaSetsCommand struct{}

type FetchReplicaSetsResult struct {
	ReplicaSets []*topo.ReplicaSet
	NodeStates  map[string]string
}

func (self *FetchReplicaSetsCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	snapshot := cs.GetClusterSnapshot()
	if snapshot == nil {
		return nil, nil
	}
	snapshot.BuildReplicaSets()

	nodeStates := map[string]string{}
	nss := cs.AllNodeStates()
	for id, n := range nss {
		nodeStates[id] = n.CurrentState()
	}
	result := FetchReplicaSetsResult{
		ReplicaSets: snapshot.ReplicaSets(),
		NodeStates:  nodeStates,
	}
	return result, nil
}
