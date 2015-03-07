package topo

import (
	"errors"
)

var (
	ErrInvalidParentId = errors.New("topo: invalid parent id, master not exist")
)

type Cluster struct {
	region      string
	nodes       []*Node
	regionNodes []*Node
	replicaSets []*ReplicaSet
	idTable     map[string]*Node
}

func NewCluster(region string) *Cluster {
	c := &Cluster{}
	c.region = region
	c.nodes = []*Node{}
	c.regionNodes = []*Node{}
	c.replicaSets = []*ReplicaSet{}
	c.idTable = map[string]*Node{}
	return c
}

func (self *Cluster) AddNode(s *Node) {
	self.idTable[s.Id()] = s
	self.nodes = append(self.nodes, s)

	if s.Region() == self.region {
		self.regionNodes = append(self.regionNodes, s)
	}
}

func (self *Cluster) Nodes() []*Node {
	return self.nodes
}

func (self *Cluster) NumNode() int {
	return len(self.nodes)
}

func (self *Cluster) RegionNodes() []*Node {
	return self.regionNodes
}

func (self *Cluster) NumRegionNode() int {
	return len(self.regionNodes)
}

func (self *Cluster) FindNode(id string) *Node {
	return self.idTable[id]
}

func (self *Cluster) Region() string {
	return self.region
}

func (self *Cluster) FailureNodes() []*Node {
	ss := []*Node{}
	for _, s := range self.regionNodes {
		if s.fail {
			ss = append(ss, s)
		}
	}
	return ss
}

func (self *Cluster) BuildReplicaSets() error {
	for _, s := range self.nodes {
		if s.IsMaster() {
			rs := NewReplicaSet()
			rs.SetMaster(s)
			self.replicaSets = append(self.replicaSets, rs)
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

	return nil
}

func (self *Cluster) String() string {
	return ""
}
