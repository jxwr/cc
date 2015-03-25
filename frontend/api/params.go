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

type MigrateParams struct {
	SourceId string   `json:"source_id"`
	TargetId string   `json:"target_id"`
	Ranges   []string `json:"ranges"`
}

type ToggleModeParams struct {
	Action string `json:"action"`
	Perm   string `json:"perm"`
	NodeId string `json:"node_id"`
}

type MakeReplicaSetParams struct {
	NodeIds []string `json:"node_ids"`
}
