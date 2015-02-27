package main

import (
	_ "fmt"
	"github.com/jxwr/cc/topo"
)

type ClusterManager struct {
	eventChan chan StateEvent
}

func NewClusterManager() *ClusterManager {
	mgr := &ClusterManager{}
	mgr.eventChan = make(chan StateEvent, 1024)
	return mgr
}

func (self *ClusterManager) PollStateEvent() (event StateEvent) {
	event = <-self.eventChan
	return
}

func (self *ClusterManager) PushStateEvent(event StateEvent) {
	self.eventChan <- event
	return
}

func (self *ClusterManager) BuildCluster() error {
	return nil
}

func (self *ClusterManager) EnableServerRead(server *topo.Server) error {
	return nil
}

func (self *ClusterManager) DisableServerRead(server *topo.Server) error {
	return nil
}

func (self *ClusterManager) HandleFailover(server *topo.Server) error {
	return nil
}

func (self *ClusterManager) AutoRebalance() error {
	return nil
}
