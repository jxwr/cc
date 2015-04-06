package inspector

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/topo"
	"github.com/jxwr/cc/utils"
)

func SendRegionTopoSnapshot(nodes []*topo.Node) error {
	params := &api.RegionSnapshotParams{
		Region:   meta.LocalRegion(),
		PostTime: time.Now().Unix(),
		Nodes:    nodes,
	}

	var resp api.MapResp
	fail, err := utils.HttpPost(api.RegionSnapshotPath, params, &resp, 30*time.Second)
	if err != nil {
		return err
	}
	if fail != nil {
		return fmt.Errorf("%d %s", fail.StatusCode, fail.Message)
	}
	return nil
}

func (self *Inspector) Run() {
	tickChan := time.NewTicker(time.Second * 1).C
	for {
		select {
		case <-tickChan:
			if !meta.IsRegionLeader() {
				continue
			}
			cluster, err := self.BuildClusterTopo()
			if err != nil {
				glog.Infof("build cluster topo failed, %v", err)
				continue
			}
			err = SendRegionTopoSnapshot(cluster.LocalRegionNodes())
			if err != nil {
				glog.Infof("send snapshot failed, %v", err)
			}
		}
	}
}
