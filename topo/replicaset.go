package topo

type ReplicaSet struct {
	master *Node
	slaves []*Node
}

func NewReplicaSet() *ReplicaSet {
	rs := &ReplicaSet{}
	return rs
}

func (s *ReplicaSet) SetMaster(node *Node) {
	s.master = node
}

func (s *ReplicaSet) Master() *Node {
	return s.master
}

func (s *ReplicaSet) AddSlave(node *Node) {
	s.slaves = append(s.slaves, node)
}

func (s *ReplicaSet) Slaves() []*Node {
	return s.slaves
}

func (s *ReplicaSet) AllNodes() []*Node {
	return append(s.slaves, s.master)
}

func (s *ReplicaSet) RegionNodes(region string) []*Node {
	nodes := []*Node{}
	for _, n := range s.AllNodes() {
		if n.Region == region {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

func (s *ReplicaSet) HasNode(nodeId string) bool {
	return s.FindNode(nodeId) != nil
}

func (s *ReplicaSet) FindNode(nodeId string) *Node {
	if nodeId == s.master.Id {
		return s.master
	}
	for _, node := range s.slaves {
		if nodeId == node.Id {
			return node
		}
	}
	return nil
}

// 该RS是否覆盖所有Region
func (s *ReplicaSet) IsCoverAllRegions(regions []string) bool {
	regionMap := map[string]bool{}
	for _, r := range regions {
		regionMap[r] = false
	}
	for _, n := range s.AllNodes() {
		regionMap[n.Region] = true
	}
	for _, v := range regionMap {
		if v == false {
			return false
		}
	}
	return true
}
