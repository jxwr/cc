package frontend

import (
	"fmt"
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

func NewFrontEnd(c *cc.Controller, httpPort, wsPort int) *FrontEnd {
	fe := &FrontEnd{
		C:            c,
		Router:       gin.Default(),
		HttpBindAddr: fmt.Sprintf(":%d", httpPort),
		WsBindAddr:   fmt.Sprintf(":%d", wsPort),
	}

	fe.Router.Static("/ui", "./public")
	fe.Router.GET(api.AppInfoPath, fe.HandleAppInfo)
	fe.Router.POST(api.RegionSnapshotPath, fe.HandleRegionSnapshot)
	fe.Router.POST(api.MigrateCreatePath, fe.HandleMigrateCreate)
	fe.Router.POST(api.RebalancePath, fe.HandleRebalance)
	fe.Router.POST(api.NodePermPath, fe.HandleToggleMode)
	fe.Router.POST(api.NodeMeetPath, fe.HandleMeetNode)
	fe.Router.POST(api.NodeForgetPath, fe.HandleForgetNode)
	fe.Router.POST(api.NodeReplicatePath, fe.HandleReplicate)
	fe.Router.POST(api.MakeReplicaSetPath, fe.HandleMakeReplicaSet)

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

func (fe *FrontEnd) HandleToggleMode(c *gin.Context) {
	var params api.ToggleModeParams
	c.Bind(&params)

	var cmd cc.Command
	nodeId := params.NodeId

	fmt.Println(params)

	if params.Action == "enable" && params.Perm == "read" {
		cmd = &command.EnableReadCommand{nodeId}
	} else if params.Action == "disable" && params.Perm == "read" {
		cmd = &command.DisableReadCommand{nodeId}
	} else if params.Action == "enable" && params.Perm == "write" {
		cmd = &command.EnableWriteCommand{nodeId}
	} else if params.Action == "disable" && params.Perm == "write" {
		cmd = &command.DisableWriteCommand{nodeId}
	} else {
		c.JSON(500, api.FailureResponse{
			Message:     "Invalid params",
			Description: fmt.Sprintf("%v", params),
		})
		return
	}

	result, err := fe.C.ProcessCommand(cmd, 2*time.Second)
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

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func (fe *FrontEnd) HandleMakeReplicaSet(c *gin.Context) {
	var params api.MakeReplicaSetParams
	c.Bind(&params)

	cmd := command.MakeReplicaSetCommand{
		NodeIds: params.NodeIds,
	}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func (fe *FrontEnd) HandleRebalance(c *gin.Context) {
	var params api.RebalanceParams
	c.Bind(&params)

	cmd := command.RebalanceCommand{
		Method:       params.Method,
		TargetIds:    params.TargetIds,
		ShowPlanOnly: params.ShowPlanOnly,
	}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func (fe *FrontEnd) HandleAppInfo(c *gin.Context) {
	cmd := command.AppInfoCommand{}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func (fe *FrontEnd) HandleMeetNode(c *gin.Context) {
	var params api.MeetNodeParams
	c.Bind(&params)

	cmd := command.MeetNodeCommand{params.NodeId}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func (fe *FrontEnd) HandleForgetNode(c *gin.Context) {
	var params api.ForgetNodeParams
	c.Bind(&params)

	cmd := command.ForgetNodeCommand{params.NodeId}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}

func (fe *FrontEnd) HandleReplicate(c *gin.Context) {
	var params api.ReplicateParams
	c.Bind(&params)

	cmd := command.ReplicateCommand{params.ChildId, params.ParentId}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(500, api.FailureResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, result)
}
