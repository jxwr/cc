package command

import (
	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/meta"
)

type AppInfoCommand struct{}

type AppInfoResult struct {
	AppConfig *meta.AppConfig
	Leader    *meta.ControllerConfig
}

func (self *AppInfoCommand) Execute(c *cc.Controller) (cc.Result, error) {
	result := &AppInfoResult{
		AppConfig: meta.GetAppConfig(),
		Leader:    meta.ClusterLeaderConfig(),
	}
	return result, nil
}
