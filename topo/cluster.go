package topo

import (
	"errors"
)

var (
	ErrInvalidParentId = errors.New("topo: invalid parent id, master not exist")
)

type Cluster struct {
	localRegion      string
	localRegionNodes []*Node
	nodes            []*Node
	replicaSets      []*ReplicaSet
	idTable          map[string]*Node
}

func NewCluster(region string) *Cluster {
	c := &Cluster{}
	c.localRegion = region
	c.localRegionNodes = []*Node{}
	c.nodes = []*Node{}
	c.replicaSets = []*ReplicaSet{}
	c.idTable = map[string]*Node{}
	return c
}

func (self *Cluster) AddNode(s *Node) {
	self.idTable[s.Id()] = s
	self.nodes = append(self.nodes, s)

	if s.Region() == self.localRegion {
		self.localRegionNodes = append(self.localRegionNodes, s)
	}
}

func (self *Cluster) Nodes() []*Node {
	return self.nodes
}

func (self *Cluster) NumNode() int {
	return len(self.nodes)
}

func (self *Cluster) LocalRegionNodes() []*Node {
	return self.localRegionNodes
}

func (self *Cluster) RegionNodes(region string) []*Node {
	nodes := []*Node{}
	for _, n := range self.nodes {
		if n.Region() == region {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

func (self *Cluster) NumLocalRegionNode() int {
	return len(self.localRegionNodes)
}

func (self *Cluster) FindNode(id string) *Node {
	return self.idTable[id]
}

func (self *Cluster) FindReplicaSetByNode(id string) *ReplicaSet {
	for _, rs := range self.replicaSets {
		if rs.HasNode(id) {
			return rs
		}
	}
	return nil
}

func (self *Cluster) Region() string {
	return self.localRegion
}

func (self *Cluster) FailureNodes() []*Node {
	ss := []*Node{}
	for _, s := range self.localRegionNodes {
		if s.fail {
			ss = append(ss, s)
		}
	}
	return ss
}

func (self *Cluster) BuildReplicaSets() error {
	replicaSets := []*ReplicaSet{}

	for _, s := range self.nodes {
		if s.IsMaster() {
			rs := NewReplicaSet()
			rs.SetMaster(s)
			replicaSets = append(replicaSets, rs)
		}
	}

	for _, s := range self.nodes {
		if !s.IsMaster() {
			master := self.FindNode(s.ParentId())
			if master == nil {
				return ErrInvalidParentId
			}

			for _, rs := range self.replicaSets {
				if rs.Master() == master {
					rs.AddSlave(s)
				}
			}
		}
	}

	self.replicaSets = replicaSets
	return nil
}

func (self *Cluster) String() string {
	return ""
}
