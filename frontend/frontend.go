package frontend

import (
	"time"

	"github.com/gin-gonic/gin"
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/frontend/api"
)

type FrontEnd struct {
	C        *cc.Controller
	Router   *gin.Engine
	BindAddr string
}

func NewFrontEnd(c *cc.Controller, bind string) *FrontEnd {
	fe := &FrontEnd{
		C:        c,
		Router:   gin.Default(),
		BindAddr: bind,
	}

	fe.Router.POST(api.RegionSnapshotPath, fe.HandleRegionSnapshot)

	return fe
}

func (fe *FrontEnd) Run() {
	fe.Router.Run(fe.BindAddr)
}

func (fe *FrontEnd) HandleRegionSnapshot(c *gin.Context) {
	var params api.RegionSnapshotParams
	c.Bind(&params)

	cmd := command.UpdateRegionCommand{
		Region: params.Region,
		Nodes:  params.Nodes,
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
