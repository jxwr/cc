package meta

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"launchpad.net/gozk"
)

type metadata struct {
	appName                string
	localIp                string
	httpPort               int
	wsPort                 int
	localRegion            string
	selfZNodeName          string
	clusterLeaderZNodeName string
	regionLeaderZNodeName  string
	appConfig              *AppConfig
	clusterLeaderConfig    *ControllerConfig
	regionLeaderConfig     *ControllerConfig
}

var theMetadata metadata

func FetchAppConfig(zconn *zookeeper.Conn, appName string) (*AppConfig, error) {
	data, _, err := zconn.Get("/r3/app/" + appName)
	if err != nil {
		return nil, err
	}
	var c AppConfig
	err = json.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, fmt.Errorf("zk: parse app config error, %v", err)
	}
	if c.AppName != appName {
		return nil, fmt.Errorf("zk: local appname is different from zk, %s <-> %s", appName, c.AppName)
	}
	if c.MasterRegion == "" {
		return nil, fmt.Errorf("zk: master region not set")
	}
	if len(c.Regions) == 0 {
		return nil, fmt.Errorf("zk: regions empty")
	}
	return &c, nil
}

func RegisterLocalController(zconn *zookeeper.Conn) error {
	ccDirPath := "/r3/app/" + theMetadata.appName + "/controller"
	zkPath := fmt.Sprintf(ccDirPath + "/cc_" + theMetadata.localRegion + "_")
	conf := &ControllerConfig{
		Ip:       theMetadata.localIp,
		HttpPort: theMetadata.httpPort,
		Region:   theMetadata.localRegion,
		WsPort:   theMetadata.wsPort,
	}
	data, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	path, err := zconn.Create(zkPath, string(data), zookeeper.SEQUENCE|zookeeper.EPHEMERAL, zookeeper.WorldACL(PERM_FILE))
	if err == nil {
		xs := strings.Split(path, "/")
		theMetadata.selfZNodeName = xs[len(xs)-1]
	}
	return err
}

func FetchLeaderConfig(zconn *zookeeper.Conn, zkPath string) (*ControllerConfig, error) {
	data, _, err := zconn.Get(zkPath)
	if err != nil {
		return nil, err
	}
	var c ControllerConfig
	err = json.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func Init(appName, localRegion string, httpPort, wsPort int, zkAddr string) error {
	theMetadata = metadata{
		appName:     appName,
		httpPort:    httpPort,
		localRegion: localRegion,
	}

	zconn, _, err := DialZk(zkAddr)
	if err != nil {
		return fmt.Errorf("zk: can't connect: %v", err)
	}

	a, err := FetchAppConfig(zconn, appName)
	if err != nil {
		return err
	}
	theMetadata.appConfig = a

	// Controller目录，如果不存在就创建
	ccDirPath := "/r3/app/" + appName + "/controller"
	CreateRecursive(zconn, ccDirPath, "", 0, zookeeper.WorldACL(zookeeper.PERM_ALL))

	err = RegisterLocalController(zconn)
	if err != nil {
		return err
	}

	clusterLeader, regionLeader, _, err := ElectLeader(zconn, ccDirPath, localRegion, false)
	if err != nil {
		return err
	}

	// 获取ClusterLeader配置
	c, err := FetchLeaderConfig(zconn, ccDirPath+"/"+clusterLeader)
	if err != nil {
		return err
	}
	theMetadata.clusterLeaderConfig = c

	// 获取RegionLeader配置
	c, err = FetchLeaderConfig(zconn, ccDirPath+"/"+regionLeader)
	if err != nil {
		return err
	}
	theMetadata.regionLeaderConfig = c

	go ZkWatcher(zconn)
	return nil
}

func LocalRegion() string {
	return theMetadata.localRegion
}

func AutoFailover() bool {
	return theMetadata.appConfig.AutoFailover
}

func LeaderHttpAddress() string {
	c := theMetadata.clusterLeaderConfig
	addr := fmt.Sprintf("%s:%d", c.Ip, c.HttpPort)
	return addr
}

func IsRegionLeader() bool {
	return theMetadata.selfZNodeName == theMetadata.regionLeaderZNodeName
}

func IsClusterLeader() bool {
	return theMetadata.selfZNodeName == theMetadata.clusterLeaderZNodeName
}

func FailoverInDoing() bool {
	return false
}

func LastFailoverTime() time.Time {
	return time.Now()
}
