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
