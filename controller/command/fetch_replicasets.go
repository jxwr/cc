package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/topo"
)

type FetchReplicaSetsCommand struct{}

type FetchReplicaSetsResult struct {
	ReplicaSets []*topo.ReplicaSet
}

func (self *FetchReplicaSetsCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	snapshot := cs.GetClusterSnapshot()
	snapshot.BuildReplicaSets()
	result := FetchReplicaSetsResult{
		ReplicaSets: snapshot.ReplicaSets(),
	}
	return result, nil
}
