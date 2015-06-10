package command

import (
	cc "github.com/jxwr/cc/controller"
)

/// Command types
func (self *EnableReadCommand) Type() cc.CommandType          { return cc.CLUSTER_COMMAND }
func (self *DisableReadCommand) Type() cc.CommandType         { return cc.CLUSTER_COMMAND }
func (self *EnableWriteCommand) Type() cc.CommandType         { return cc.CLUSTER_COMMAND }
func (self *DisableWriteCommand) Type() cc.CommandType        { return cc.CLUSTER_COMMAND }
func (self *MakeReplicaSetCommand) Type() cc.CommandType      { return cc.CLUSTER_COMMAND }
func (self *ForgetAndResetNodeCommand) Type() cc.CommandType  { return cc.CLUSTER_COMMAND }
func (self *FailoverBeginCommand) Type() cc.CommandType       { return cc.CLUSTER_COMMAND }
func (self *FetchReplicaSetsCommand) Type() cc.CommandType    { return cc.CLUSTER_COMMAND }
func (self *FailoverTakeoverCommand) Type() cc.CommandType    { return cc.CLUSTER_COMMAND }
func (self *MeetNodeCommand) Type() cc.CommandType            { return cc.CLUSTER_COMMAND }
func (self *ReplicateCommand) Type() cc.CommandType           { return cc.CLUSTER_COMMAND }
func (self *MigrateCommand) Type() cc.CommandType             { return cc.CLUSTER_COMMAND }
func (self *MigratePauseCommand) Type() cc.CommandType        { return cc.CLUSTER_COMMAND }
func (self *MigrateResumeCommand) Type() cc.CommandType       { return cc.CLUSTER_COMMAND }
func (self *MigrateCancelCommand) Type() cc.CommandType       { return cc.CLUSTER_COMMAND }
func (self *SetAsMasterCommand) Type() cc.CommandType         { return cc.CLUSTER_COMMAND }
func (self *UpdateRegionCommand) Type() cc.CommandType        { return cc.CLUSTER_COMMAND }
func (self *RebalanceCommand) Type() cc.CommandType           { return cc.CLUSTER_COMMAND }
func (self *FetchMigrationTasksCommand) Type() cc.CommandType { return cc.CLUSTER_COMMAND }
func (self *MergeSeedsCommand) Type() cc.CommandType          { return cc.REGION_COMMAND }
