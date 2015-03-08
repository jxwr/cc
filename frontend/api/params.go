package api

import (
	"github.com/jxwr/cc/topo"
)

type FailureResponse struct {
	StatusCode  int    `json:"status_code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

type MapResp map[string]interface{}

type RegionSnapshotParams struct {
	Region   string       `json:"region"`
	PostTime int64        `json:"posttime"`
	Nodes    []*topo.Node `json:"nodes"`
}
