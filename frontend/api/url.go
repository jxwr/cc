package api

const (
	// URL值得仔细设计一番
	AppInfoPath             = "/app/info"
	RegionSnapshotPath      = "/region/snapshot"
	MergeSeedsPath          = "/region/mergeseeds"
	MigrateCreatePath       = "/migrate/create"
	MigratePausePath        = "/migrate/pause"
	MigrateResumePath       = "/migrate/resume"
	MigrateCancelPath       = "/migrate/cancel"
	FetchMigrationTasksPath = "/migrate/tasks"
	RebalancePath           = "/migrate/rebalance"
	NodePermPath            = "/node/perm"
	NodeMeetPath            = "/node/meet"
	NodeForgetAndResetPath  = "/node/forgetAndReset"
	NodeReplicatePath       = "/node/replicate"
	NodeResetPath           = "/node/reset"
	NodeSetAsMasterPath     = "/node/setAsMaster"
	FetchReplicaSetsPath    = "/replicasets"
	MakeReplicaSetPath      = "/replicaset/make"
	FailoverTakeoverPath    = "/failover/takeover"
	LogSlicePath            = "/log/slice"
)
