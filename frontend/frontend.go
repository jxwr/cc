package frontend

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/controller/command"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/frontend/auth"
	"github.com/ksarch-saas/cc/log"
	"github.com/ksarch-saas/cc/topo"
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
	store := auth.NewTokenStore("r3")
	tokenAuth := auth.NewTokenAuth(nil, store, nil)

	fe.Router.Static("/ui", "./public")
	fe.Router.GET(api.AppInfoPath, fe.HandleAppInfo)
	fe.Router.GET(api.FetchReplicaSetsPath, fe.HandleFetchReplicaSets)
	fe.Router.POST(api.LogSlicePath, fe.HandleLogSlice)
	fe.Router.POST(api.RegionSnapshotPath, fe.HandleRegionSnapshot)
	fe.Router.POST(api.MigrateCreatePath, tokenAuth.HandleFunc(fe.HandleMigrateCreate))
	fe.Router.POST(api.MigratePausePath, tokenAuth.HandleFunc(fe.HandleMigratePause))
	fe.Router.POST(api.MigrateResumePath, tokenAuth.HandleFunc(fe.HandleMigrateResume))
	fe.Router.POST(api.MigrateCancelPath, tokenAuth.HandleFunc(fe.HandleMigrateCancel))
	fe.Router.GET(api.FetchMigrationTasksPath, fe.HandleFetchMigrationTasks)
	fe.Router.POST(api.RebalancePath, tokenAuth.HandleFunc(fe.HandleRebalance))
	fe.Router.POST(api.NodePermPath, tokenAuth.HandleFunc(fe.HandleToggleMode))
	fe.Router.POST(api.NodeMeetPath, tokenAuth.HandleFunc(fe.HandleMeetNode))
	fe.Router.POST(api.NodeSetAsMasterPath, tokenAuth.HandleFunc(fe.HandleSetAsMaster))
	fe.Router.POST(api.NodeForgetAndResetPath, tokenAuth.HandleFunc(fe.HandleForgetAndResetNode))
	fe.Router.POST(api.NodeReplicatePath, tokenAuth.HandleFunc(fe.HandleReplicate))
	fe.Router.POST(api.MakeReplicaSetPath, tokenAuth.HandleFunc(fe.HandleMakeReplicaSet))
	fe.Router.POST(api.FailoverTakeoverPath, tokenAuth.HandleFunc(fe.HandleFailoverTakeover))
	fe.Router.POST(api.MergeSeedsPath, fe.HandleMergeSeeds)

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
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleToggleMode(c *gin.Context) {
	var params api.ToggleModeParams
	c.Bind(&params)

	var cmd cc.Command
	nodeId := params.NodeId

	if params.Action == "enable" && params.Perm == "read" {
		cmd = &command.EnableReadCommand{nodeId}
	} else if params.Action == "disable" && params.Perm == "read" {
		cmd = &command.DisableReadCommand{nodeId}
	} else if params.Action == "enable" && params.Perm == "write" {
		cmd = &command.EnableWriteCommand{nodeId}
	} else if params.Action == "disable" && params.Perm == "write" {
		cmd = &command.DisableWriteCommand{nodeId}
	} else {
		c.JSON(200, api.MakeFailureResponse("Invalid command"))
		return
	}

	result, err := fe.C.ProcessCommand(cmd, 2*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
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
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleMigratePause(c *gin.Context) {
	var params api.MigrateActionParams
	c.Bind(&params)

	cmd := command.MigratePauseCommand{
		SourceId: params.SourceId,
	}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleMigrateResume(c *gin.Context) {
	var params api.MigrateActionParams
	c.Bind(&params)

	cmd := command.MigrateResumeCommand{
		SourceId: params.SourceId,
	}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleMigrateCancel(c *gin.Context) {
	var params api.MigrateActionParams
	c.Bind(&params)

	cmd := command.MigrateCancelCommand{
		SourceId: params.SourceId,
	}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleMakeReplicaSet(c *gin.Context) {
	var params api.MakeReplicaSetParams
	c.Bind(&params)

	cmd := command.MakeReplicaSetCommand{
		NodeIds: params.NodeIds,
	}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
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
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleAppInfo(c *gin.Context) {
	cmd := command.AppInfoCommand{}

	result, err := cmd.Execute(fe.C)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleFetchReplicaSets(c *gin.Context) {
	cmd := command.FetchReplicaSetsCommand{}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleFetchMigrationTasks(c *gin.Context) {
	cmd := command.FetchMigrationTasksCommand{}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleMeetNode(c *gin.Context) {
	var params api.MeetNodeParams
	c.Bind(&params)

	cmd := command.MeetNodeCommand{params.NodeId}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleSetAsMaster(c *gin.Context) {
	var params api.SetAsMasterParams
	c.Bind(&params)

	cmd := command.SetAsMasterCommand{params.NodeId}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleForgetAndResetNode(c *gin.Context) {
	var params api.ForgetAndResetNodeParams
	c.Bind(&params)

	cmd := command.ForgetAndResetNodeCommand{params.NodeId}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleReplicate(c *gin.Context) {
	var params api.ReplicateParams
	c.Bind(&params)

	cmd := command.ReplicateCommand{params.ChildId, params.ParentId}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleFailoverTakeover(c *gin.Context) {
	var params api.FailoverTakeoverParams
	c.Bind(&params)

	cmd := command.FailoverTakeoverCommand{params.NodeId}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleMergeSeeds(c *gin.Context) {
	var params api.MergeSeedsParams
	c.Bind(&params)

	cmd := command.MergeSeedsCommand{params.Region, params.Seeds}

	result, err := fe.C.ProcessCommand(&cmd, 5*time.Second)
	if err != nil {
		c.JSON(200, api.MakeFailureResponse(err.Error()))
		return
	}

	c.JSON(200, api.MakeSuccessResponse(result))
}

func (fe *FrontEnd) HandleLogSlice(c *gin.Context) {
	var params api.LogSliceParams
	c.Bind(&params)

	n := len(log.LogRingBuffer)
	lines := log.LogRingBuffer[n-params.Pos-params.Count : n-params.Pos]

	c.JSON(200, api.MakeSuccessResponse(lines))
}
