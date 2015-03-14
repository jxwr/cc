package frontend

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/topo"
)

type FrontEnd struct {
	C            *cc.Controller
	Router       *gin.Engine
	HttpBindAddr string
	WsBindAddr   string
}

func NewFrontEnd(c *cc.Controller, httpBind, wsBind string) *FrontEnd {
	fe := &FrontEnd{
		C:            c,
		Router:       gin.Default(),
		HttpBindAddr: httpBind,
		WsBindAddr:   wsBind,
	}

	fe.Router.Static("/ui", "./public")
	fe.Router.POST(api.RegionSnapshotPath, fe.HandleRegionSnapshot)
	fe.Router.POST(api.MigrateCreatePath, fe.HandleMigrateCreate)

	return fe
}

func (fe *FrontEnd) Run() {
	go RunWebsockServer(fe.WsBindAddr)
	fe.Router.Run(fe.HttpBindAddr)
}

func (fe *FrontEnd) HandleRegionSnapshot(c *gin.Context) {
	var params api.RegionSnapshotParams
	c.Bind(&params)

	cmd := command.UpdateRegionCommand{
		Region: params.Region,
		Nodes:  params.Nodes,
	}

	result, err := fe.C.ProcessCommand(&cmd, 2*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func (fe *FrontEnd) HandleMigrateCreate(c *gin.Context) {
	var params api.MigrateParams
	c.Bind(&params)

	ranges := []topo.Range{}
	for _, r := range params.Ranges {
		xs := strings.Split(r, "-")
		if len(xs) == 2 {
			left, _ := strconv.Atoi(xs[0])
			right, _ := strconv.Atoi(xs[1])
			ranges = append(ranges, topo.Range{left, right})
		} else {
			left, _ := strconv.Atoi(r)
			ranges = append(ranges, topo.Range{left, left})
		}
	}

	cmd := command.MigrateCommand{
		SourceId: params.SourceId,
		TargetId: params.TargetId,
		Ranges:   ranges,
	}

	result, err := fe.C.ProcessCommand(&cmd, 2*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}
