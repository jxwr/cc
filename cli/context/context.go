package context

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/utils"
	zookeeper "github.com/samuel/go-zookeeper/zk"
)

var (
	appConfig               meta.AppConfig
	controllerConfig        meta.ControllerConfig
	nodesCache              []string
	ErrNoNodesFound         = errors.New("Command failed: no nodes found")
	ErrMoreThanOneNodeFound = errors.New("Command failed: more than one node found")
	ZkAddr                  string
)

func SetApp(appName string, zkAddr string) error {
	zconn, _, err := meta.DialZk(zkAddr)
	defer zconn.Close()
	if err != nil {
		return fmt.Errorf("zk: can't connect: %v", err)
	}

	// get 1st controller
	children, _, err := zconn.Children("/r3/app/" + appName + "/controller")
	if len(children) == 0 {
		return fmt.Errorf("no controller found")
	}
	data, _, err := zconn.Get("/r3/app/" + appName + "/controller/" + children[0])
	var cc meta.ControllerConfig
	err = json.Unmarshal([]byte(data), &cc)
	if err != nil {
		return err
	}
	// fetch app info
	url := fmt.Sprintf("http://%s:%d"+api.AppInfoPath, cc.Ip, cc.HttpPort)
	resp, err := utils.HttpGet(url, nil, 5*time.Second)
	if err != nil {
		return err
	}
	// map to structure
	var res command.AppInfoResult
	err = utils.InterfaceToStruct(resp.Body, &res)
	if err != nil {
		return err
	}
	appConfig = *res.AppConfig
	controllerConfig = *res.Leader

	fmt.Fprintf(os.Stderr, "[ leader : %s:%d ]\n", controllerConfig.Ip, controllerConfig.HttpPort)
	fmt.Fprintf(os.Stderr, "[ web    : http://%s:%d/ui/cluster.html ]\n",
		controllerConfig.Ip, controllerConfig.HttpPort)
	err = CacheNodes()
	return err
}

func AddApp(appName string, config []byte) error {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer zconn.Close()
	if err != nil {
		return fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/app/" + appName
	exists, _, err := zconn.Exists(zkPath)
	if err != nil {
		return fmt.Errorf("zk: call exist failed %v", err)
	}
	if exists {
		return fmt.Errorf("zk: %s node already exists", appName)
	} else {
		//add node
		_, err := zconn.Create(zkPath, config, 0, zookeeper.WorldACL(zookeeper.PermAll))
		if err != nil {
			return fmt.Errorf("zk: create failed %v", err)
		}
		return nil
	}
}

func ModApp(appName string, config []byte, version int32) error {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer zconn.Close()
	if err != nil {
		return fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/app/" + appName
	exists, _, err := zconn.Exists(zkPath)
	if err != nil {
		return fmt.Errorf("zk: call exist failed %v", err)
	}
	if !exists {
		return fmt.Errorf("zk: %s node not exists", appName)
	} else {
		//update node
		_, err := zconn.Set(zkPath, config, version)
		if err != nil {
			return fmt.Errorf("zk: set failed %v", err)
		}
		return nil
	}
}

func GetApp(appName string) ([]byte, int32, error) {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer zconn.Close()
	if err != nil {
		return nil, 0, fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/app/" + appName
	config, stat, err := zconn.Get(zkPath)
	if err != nil {
		return nil, 0, fmt.Errorf("zk: get: %v", err)
	}
	return config, stat.Version, nil
}

func DelApp(appName string, version int32) error {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer zconn.Close()
	if err != nil {
		return fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/app/" + appName
	err = zconn.Delete(zkPath, version)
	if err != nil {
		return fmt.Errorf("zk: path delete %v", err)
	}
	return nil
}

func GetLeaderAddr() string {
	return fmt.Sprintf("%s:%d", controllerConfig.Ip, controllerConfig.HttpPort)
}

func GetLeaderWebSocketAddr() string {
	return fmt.Sprintf("%s:%d", controllerConfig.Ip, controllerConfig.WsPort)
}

func GetAppInfo() string {
	var data []byte
	data, _ = json.Marshal(appConfig)
	var out bytes.Buffer
	json.Indent(&out, []byte(data), "", "  ")

	return out.String()
}

func CacheNodes() error {
	addr := GetLeaderAddr()
	url := "http://" + addr + api.FetchReplicaSetsPath

	resp, err := utils.HttpGet(url, nil, 5*time.Second)
	if err != nil {
		return err
	}

	var rss command.FetchReplicaSetsResult
	err = utils.InterfaceToStruct(resp.Body, &rss)
	if err != nil {
		return err
	}
	for _, rs := range rss.ReplicaSets {
		nodes := rs.AllNodes()
		for _, node := range nodes {
			nodesCache = append(nodesCache, node.Id)
		}
	}
	return nil
}

func GetId(shortid string) (string, error) {
	var longid string
	cnt := 0
	for _, node := range nodesCache {
		if strings.HasPrefix(node, shortid) {
			longid = node
			cnt = cnt + 1
		}
	}
	if cnt == 0 {
		return "", ErrNoNodesFound
	} else if cnt > 1 {
		return "", ErrMoreThanOneNodeFound
	} else {
		return longid, nil
	}
}
