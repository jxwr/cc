package context

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/jxwr/cc/meta"
)

var appConfig meta.AppConfig
var controllerConfig meta.ControllerConfig

func SetApp(appName string, zkAddr string) error {
	zconn, _, err := meta.DialZk(zkAddr)
	if err != nil {
		return fmt.Errorf("zk: can't connect: %v", err)
	}

	// fetch app config
	data, _, err := zconn.Get("/r3/app/" + appName)
	if err != nil {
		return err
	}
	var ac meta.AppConfig
	err = json.Unmarshal([]byte(data), &ac)
	if err != nil {
		return fmt.Errorf("parse app config error, %v", err)
	}
	appConfig = ac

	// show config
	var aout bytes.Buffer
	json.Indent(&aout, []byte(data), "", "  ")
	fmt.Println("AppConfig:\n", aout.String())

	// fetch leader
	children, _, err := zconn.Children("/r3/app/" + appName + "/controller")
	if len(children) == 0 {
		return fmt.Errorf("no controller found")
	}
	data, _, err = zconn.Get("/r3/app/" + appName + "/controller/" + children[0])
	var cc meta.ControllerConfig
	err = json.Unmarshal([]byte(data), &cc)
	if err != nil {
		return err
	}
	controllerConfig = cc

	// show config
	var cout bytes.Buffer
	json.Indent(&cout, []byte(data), "", "  ")
	fmt.Println("Leader:\n", cout.String())

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
