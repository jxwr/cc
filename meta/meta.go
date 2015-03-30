package meta

import (
	"fmt"
	"log"
	"time"

	"launchpad.net/gozk"
)

type Meta struct {
	/// local config
	appName     string
	localIp     string
	httpPort    int
	wsPort      int
	localRegion string

	/// leadership
	selfZNodeName          string
	clusterLeaderZNodeName string
	regionLeaderZNodeName  string

	/// /r3/app/<appname>/controller
	ccDirPath string

	/// configs in ZK
	appConfig           *AppConfig
	clusterLeaderConfig *ControllerConfig
	regionLeaderConfig  *ControllerConfig

	/// zk connection
	zconn    *zookeeper.Conn
	zsession <-chan zookeeper.Event
}

var meta *Meta

func GetAppConfig() *AppConfig {
	return meta.appConfig
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
	return meta.appConfig.MasterRegion
}

func AllRegions() []string {
	return meta.appConfig.Regions
}

func AutoFailover() bool {
	return meta.appConfig.AutoFailover
}

func LeaderHttpAddress() string {
	c := meta.clusterLeaderConfig
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

func Run(appName, localRegion string, httpPort, wsPort int, zkAddr string, initCh chan error) {
	zconn, session, err := DialZk(zkAddr)
	if err != nil {
		initCh <- fmt.Errorf("zk: can't connect: %v", err)
		return
	}

	meta = &Meta{
		appName:     appName,
		wsPort:      wsPort,
		httpPort:    httpPort,
		localRegion: localRegion,
		ccDirPath:   "/r3/app/" + appName + "/controller",
		zconn:       zconn,
		zsession:    session,
	}

	a, w, err := meta.FetchAppConfig()
	if err != nil {
		initCh <- err
		return
	}
	meta.appConfig = a
	go meta.handleAppConfigChanged(w)

	// Controller目录，如果不存在就创建
	CreateRecursive(zconn, meta.ccDirPath, "", 0, zookeeper.WorldACL(zookeeper.PERM_ALL))

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
	// 元信息初始化成功，通知Main函数继续初始化
	initCh <- nil

	// 开始各种Watch
	tickChan := time.NewTicker(time.Second * 60).C
	for {
		select {
		case event := <-meta.zsession:
			if event.State == zookeeper.STATE_EXPIRED_SESSION {
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
				log.Println("Leader election error,", err)
			}
		case <-tickChan:
			clusterLeader, regionLeader, _, err := meta.CheckLeaders(false)
			log.Println("Check leaders,", err)
			needElect := false
			if clusterLeader == "" || regionLeader == "" {
				log.Println("Leaders gone, will reelect leaders.")
				needElect = true
			} else if ClusterLeaderZNodeName() != clusterLeader {
				log.Println("Cluster leader changed, reelect.")
				needElect = true
			} else if RegionLeaderZNodeName() != regionLeader {
				log.Println("Region leader changed, reelect.")
				needElect = true
			}
			if needElect {
				watcher, err = meta.ElectLeader()
				if err != nil {
					log.Println("Leader election error,", err)
				}
			}
		}
	}
}
