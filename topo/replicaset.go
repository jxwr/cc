package topo

type ReplicaSet struct {
	master *Server
	slaves []*Server
}

func NewReplicaSet() *ReplicaSet {
	rs := &ReplicaSet{}
	return rs
}

func (s *ReplicaSet) SetMaster(server *Server) {
	s.master = server
}

func (s *ReplicaSet) Master() *Server {
	return s.master
}

func (s *ReplicaSet) AddSlave(server *Server) {
	s.slaves = append(s.slaves, server)
}

func (s *ReplicaSet) Slaves() []*Server {
	return s.slaves
}
