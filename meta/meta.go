package meta

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/topo"
	"github.com/ksarch-saas/cc/utils"
	"github.com/ksarch-saas/cc/utils/net"
	zookeeper "github.com/samuel/go-zookeeper/zk"
)

var meta *Meta

type Meta struct {
	/// local config
	appName     string
	localIp     string
	httpPort    int
	wsPort      int
	localRegion string

	/// Seed nodes
	seeds []*topo.Node

	/// leadership
	selfZNodeName          string
	clusterLeaderZNodeName string
	regionLeaderZNodeName  string

	/// /r3/app/<appname>/controller
	ccDirPath string

	/// configs in ZK
	appConfig           atomic.Value // *AppConfig
	clusterLeaderConfig *ControllerConfig
	regionLeaderConfig  *ControllerConfig

	/// zk connection
	zconn    *zookeeper.Conn
	zsession <-chan zookeeper.Event
}

func (self *Meta) HasSeed(seed *topo.Node) bool {
	for _, s := range self.seeds {
		if s.Addr() == seed.Addr() {
			if s.Id == "" {
				*s = *seed
			}
			return true
		}
	}
	return false
}

func MergeSeeds(seeds []*topo.Node) {
	for _, seed := range seeds {
		if !meta.HasSeed(seed) {
			meta.seeds = append(meta.seeds, seed)
		}
	}
}

func Seeds() []*topo.Node {
	return meta.seeds
}

func GetAppConfig() *AppConfig {
	return meta.appConfig.Load().(*AppConfig)
}

func ClusterLeaderConfig() *ControllerConfig {
	return meta.clusterLeaderConfig
}

func AppName() string {
	return meta.appName
}

func LocalRegion() string {
	return meta.localRegion
}

func MasterRegion() string {
	return GetAppConfig().MasterRegion
}

func IsInMasterRegion() bool {
	return LocalRegion() == MasterRegion()
}

func AllRegions() []string {
	return GetAppConfig().Regions
}

func AutoFailover() bool {
	return GetAppConfig().AutoFailover
}

func LeaderHttpAddress() string {
	c := meta.clusterLeaderConfig
	addr := fmt.Sprintf("%s:%d", c.Ip, c.HttpPort)
	return addr
}

func RegionLeaderHttpAddress() string {
	c := meta.regionLeaderConfig
	addr := fmt.Sprintf("%s:%d", c.Ip, c.HttpPort)
	return addr
}

func IsRegionLeader() bool {
	return meta.selfZNodeName == meta.regionLeaderZNodeName
}

func IsClusterLeader() bool {
	return meta.selfZNodeName == meta.clusterLeaderZNodeName
}

func ClusterLeaderZNodeName() string {
	return meta.clusterLeaderZNodeName
}

func RegionLeaderZNodeName() string {
	return meta.regionLeaderZNodeName
}

func IsDoingFailover() (bool, error) {
	return meta.IsDoingFailover()
}

func LastFailoverTime() (*time.Time, error) {
	r, err := meta.LastFailoverRecord()
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return &r.Timestamp, nil
}

func AddFailoverRecord(record *FailoverRecord) error {
	return meta.AddFailoverRecord(record)
}

func MarkFailoverDoing(record *FailoverRecord) error {
	return meta.MarkFailoverDoing(record)
}

func UnmarkFailoverDoing() error {
	return meta.UnmarkFailoverDoing()
}

func Run(appName, localRegion string, httpPort, wsPort int, zkAddr string, seeds []*topo.Node, initCh chan error) {
	zconn, session, err := DialZk(zkAddr)
	if err != nil {
		initCh <- fmt.Errorf("zk: can't connect: %v", err)
		return
	}

	localIp, err := net.LocalIP()
	if err != nil {
		glog.Info("meta: can not get local ip", err)
	}

	meta = &Meta{
		appName:     appName,
		wsPort:      wsPort,
		httpPort:    httpPort,
		localRegion: localRegion,
		localIp:     localIp,
		seeds:       seeds,
		ccDirPath:   "/r3/app/" + appName + "/controller",
		zconn:       zconn,
		zsession:    session,
	}

	a, w, err := meta.FetchAppConfig()
	if err != nil {
		initCh <- err
		return
	}
	meta.appConfig.Store(a)
	go meta.handleAppConfigChanged(w)

	// Controller目录，如果不存在就创建
	CreateRecursive(zconn, meta.ccDirPath, "", 0, zookeeper.WorldACL(zookeeper.PermAll))

	err = meta.RegisterLocalController()
	if err != nil {
		initCh <- err
		return
	}

	watcher, err := meta.ElectLeader()
	if err != nil {
		initCh <- err
		return
	}
	PostSeeds()
	// 元信息初始化成功，通知Main函数继续初始化
	initCh <- nil

	// 开始各种Watch
	tickChan := time.NewTicker(time.Second * 60).C
	for {
		select {
		case event := <-meta.zsession:
			if event.State == zookeeper.StateExpired {
				// 重试连接直到成功
				for {
					zconn, session, err := DialZk(zkAddr)
					if err == nil {
						meta.zconn = zconn
						meta.zsession = session
						break
					}
					time.Sleep(10 * time.Second)
				}
			}
		case <-watcher:
			watcher, err = meta.ElectLeader()
			if err != nil {
				glog.Warning("Leader election error,", err)
			}
		case <-tickChan:
			clusterLeader, regionLeader, _, err := meta.CheckLeaders(false)
			glog.Info("Check leaders, err: ", err)
			needElect := false
			if clusterLeader == "" || regionLeader == "" {
				glog.Warning("Leaders gone, will reelect leaders.")
				needElect = true
			} else if ClusterLeaderZNodeName() != clusterLeader {
				glog.Warning("Cluster leader changed, reelect.")
				needElect = true
			} else if RegionLeaderZNodeName() != regionLeader {
				glog.Warning("Region leader changed, reelect.")
				needElect = true
			}
			if needElect {
				watcher, err = meta.ElectLeader()
				if err != nil {
					glog.Warning("Leader election error,", err)
				}
			}
		}
		PostSeeds()
	}
}

func PostSeeds() {
	if !IsRegionLeader() {
		url := "http://" + RegionLeaderHttpAddress() + api.MergeSeedsPath
		req := api.MergeSeedsParams{
			Region: LocalRegion(),
			Seeds:  meta.seeds,
		}

		glog.Warningf("Post %s seeds %v to be merged", LocalRegion(), meta.seeds)

		utils.HttpPost(url, req, 5*time.Second)
	}
}
