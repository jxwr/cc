package topo

import (
	"errors"
)

var (
	ErrInvalidParentId = errors.New("topo: invalid parent id, master not exist")
)

type Cluster struct {
	region        string
	servers       []*Server
	regionServers []*Server
	replicaSets   []*ReplicaSet
	idTable       map[string]*Server
}

func NewCluster(region string) *Cluster {
	c := &Cluster{}
	c.region = region
	c.servers = []*Server{}
	c.regionServers = []*Server{}
	c.replicaSets = []*ReplicaSet{}
	c.idTable = map[string]*Server{}
	return c
}

func (self *Cluster) AddServer(s *Server) {
	self.idTable[s.Id()] = s
	self.servers = append(self.servers, s)

	if s.Region() == self.region {
		self.regionServers = append(self.regionServers, s)
	}
}

func (self *Cluster) Servers() []*Server {
	return self.servers
}

func (self *Cluster) NumServer() int {
	return len(self.servers)
}

func (self *Cluster) RegionServers() []*Server {
	return self.regionServers
}

func (self *Cluster) NumRegionServer() int {
	return len(self.regionServers)
}

func (self *Cluster) FindServer(id string) *Server {
	return self.idTable[id]
}

func (self *Cluster) FailureServers() []*Server {
	ss := []*Server{}
	for _, s := range self.regionServers {
		if s.fail {
			ss = append(ss, s)
		}
	}
	return ss
}

func (self *Cluster) BuildReplicaSets() error {
	for _, s := range self.servers {
		if s.IsMaster() {
			rs := NewReplicaSet()
			rs.SetMaster(s)
			self.replicaSets = append(self.replicaSets, rs)
		}
	}

	for _, s := range self.servers {
		if !s.IsMaster() {
			master := self.FindServer(s.ParentId())
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
