package meta

import (
	"fmt"
	"strconv"
	"strings"

	"launchpad.net/gozk"
)

func ElectLeader(zconn *zookeeper.Conn, zkPath, region string, watch bool) (string, string, <-chan zookeeper.Event, error) {
	var err error
	var children []string
	var watcher <-chan zookeeper.Event
	var stat *zookeeper.Stat

	if watch {
		children, stat, watcher, err = zconn.ChildrenW(zkPath)
	} else {
		children, stat, err = zconn.Children(zkPath)
	}
	if err != nil {
		return "", "", nil, err
	}
	if stat.NumChildren() == 0 {
		return "", "", nil, fmt.Errorf("zk: no node in controller leader directory")
	}

	needRejoin := true
	clusterMinSeq := -1
	clusterLeader := ""
	regionMinSeq := -1
	regionLeader := ""

	for _, child := range children {
		xs := strings.Split(child, "_")
		seq, _ := strconv.Atoi(xs[2])
		region := xs[1]
		// Cluster Leader
		if clusterMinSeq < 0 {
			clusterMinSeq = seq
			clusterLeader = child
		}
		if seq < clusterMinSeq {
			clusterMinSeq = seq
			clusterLeader = child
		}
		// Region Leader
		if theMetadata.localRegion == region {
			if regionMinSeq < 0 {
				regionMinSeq = seq
				regionLeader = child
			}
			if seq < regionMinSeq {
				regionMinSeq = seq
				regionLeader = child
			}
		}
		// Rejoin
		if theMetadata.selfZNodeName == child {
			needRejoin = false
		}
	}

	if needRejoin {
		err := RegisterLocalController(zconn)
		if err != nil {
			return "", "", nil, err
		}
	}

	return clusterLeader, regionLeader, watcher, nil
}
