package context

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/utils"
)

var (
	appConfig               meta.AppConfig
	controllerConfig        meta.ControllerConfig
	nodesCache              []string
	ErrNoNodesFound         = errors.New("Command failed: no nodes found")
	ErrMoreThanOneNodeFound = errors.New("Command failed: more than one node found")
)

func SetApp(appName string, zkAddr string) error {
	zconn, _, err := meta.DialZk(zkAddr)
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

	fmt.Printf("[ leader : %s:%d ]\n", controllerConfig.Ip, controllerConfig.HttpPort)
	err = CacheNodes()
	return err
}

func GetLeaderAddr() string {
	return fmt.Sprintf("%s:%d", controllerConfig.Ip, controllerConfig.HttpPort)
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
