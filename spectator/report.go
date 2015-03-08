package spectator

import (
	"fmt"
	"log"
	"time"

	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/topo"
	"github.com/jxwr/cc/util"
)

func SendRegionTopoSnapshot(nodes []*topo.Node) error {
	params := &api.RegionSnapshotParams{
		Region:   meta.LocalRegion(),
		PostTime: time.Now().Unix(),
		Nodes:    nodes,
	}

	var resp api.MapResp
	fail, err := util.HttpPost(api.RegionSnapshotPath, params, &resp, 30*time.Second)
	if err != nil {
		return err
	}
	if fail != nil {
		return fmt.Errorf("%d %s", fail.StatusCode, fail.Message)
	}
	return nil
}

func (self *Spectator) ReportRegionSnanshotLoop() {
	tickChan := time.NewTicker(time.Second * 1).C
	for {
		select {
		case <-tickChan:
			cluster, err := self.BuildClusterTopo()
			if err != nil {
				log.Println("build cluster topo failed,", err)
				continue
			}
			err = SendRegionTopoSnapshot(cluster.LocalRegionNodes())
			if err != nil {
				log.Println("send snapshot failed,", err)
			}
		}
	}
}
