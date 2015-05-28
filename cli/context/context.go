package context

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/utils"
)

var appConfig meta.AppConfig
var controllerConfig meta.ControllerConfig

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
	utils.InterfaceToStruct(resp.Body, &res)
	if err != nil {
		return err
	}
	appConfig = *res.AppConfig
	controllerConfig = *res.Leader

	fmt.Printf("[ leader : %s:%d ]\n", controllerConfig.Ip, controllerConfig.HttpPort)
	return nil
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
