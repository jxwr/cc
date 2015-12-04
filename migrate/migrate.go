package migrate

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ksarch-saas/cc/log"
	"github.com/ksarch-saas/cc/meta"
	"github.com/ksarch-saas/cc/redis"
	"github.com/ksarch-saas/cc/streams"
	"github.com/ksarch-saas/cc/topo"
)

const (
	StateRunning int32 = iota
	StatePausing
	StatePaused
	StateCancelling
	StateCancelled
	StateDone
	StateTargetNodeFailure
)

var stateNames = map[int32]string{
	StateRunning:           "Migrating",
	StatePausing:           "Pausing",
	StatePaused:            "Paused",
	StateCancelling:        "Cancelling",
	StateCancelled:         "Cancelled",
	StateDone:              "Done",
	StateTargetNodeFailure: "TargetNodeFailure",
}

type MigratePlan struct {
	SourceId string
	TargetId string
	Ranges   []topo.Range
	CurrSlot int
	State    string
	task     *MigrateTask
}

type MigrateTask struct {
	cluster          *topo.Cluster
	ranges           []topo.Range
	source           atomic.Value
	target           atomic.Value
	currRangeIndex   int // current range index
	currSlot         int // current slot
	state            int32
	backupReplicaSet *topo.ReplicaSet
	lastPubTime      time.Time
	totalKeysInSlot  int // counter of total keys migrated
}

func NewMigrateTask(cluster *topo.Cluster, sourceRS, targetRS *topo.ReplicaSet, ranges []topo.Range) *MigrateTask {
	t := &MigrateTask{
		cluster:     cluster,
		ranges:      ranges,
		state:       StateRunning,
		lastPubTime: time.Now(),
	}
	t.ReplaceSourceReplicaSet(sourceRS)
	t.ReplaceTargetReplicaSet(targetRS)
	return t
}

func (t *MigrateTask) TaskName() string {
	return fmt.Sprintf("Mig(%s_To_%s)", t.SourceNode().Id[:6], t.TargetNode().Id[:6])
}

func (t *MigrateTask) ToPlan() *MigratePlan {
	return &MigratePlan{
		SourceId: t.SourceNode().Id,
		TargetId: t.TargetNode().Id,
		Ranges:   t.ranges,
		CurrSlot: t.currSlot,
		State:    stateNames[t.state],
	}
}

/// 迁移slot过程:
/// 1. 标记Target分片Master为IMPORTING
/// 2. 标记所有Source分片节点为MIGRATING
/// 3. 从Source分片Master取keys迁移，直到空，数据迁移完成
/// 4. 设置Target的Slave的slot归属到Target
/// 5. 设置Target的Master的slot归属到Target
/// 6. 设置Source所有节点的slot归属到Target
/// 命令:
/// 1. <Target Master> setslot $slot IMPORTING $sourceId
/// 2. <Source Slaves> setslot $slot MIGRATING $targetId
/// 3. <Source Master> setslot $slot MIGRATING $targetId
/// ... migrating all keys
/// 4. <Target Slaves> setslot $slot node $targetId
/// 5. <Target Master> setslot $slot node $targetId
/// 6. <Source Slaves> setslot $slot node $targetId
/// 7. <Source Master> setslot $slot node $targetId
func (t *MigrateTask) migrateSlot(slot int, keysPer int) (int, error, string) {
	rs := t.SourceReplicaSet()
	sourceNode := t.SourceNode()
	targetNode := t.TargetNode()

	err := redis.SetSlot(targetNode.Addr(), slot, redis.SLOT_IMPORTING, sourceNode.Id)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ERR I'm already the owner of hash slot") {
			log.Warningf(t.TaskName(), "%s already the owner of hash slot %d",
				targetNode.Id[:6], slot)
			// 逻辑到此，说明Target已经包含该slot，但是Source处于Migrating状态
			// 迁移实际已经完成，需要清理Source的Migrating状态
			srs := t.SourceReplicaSet()
			err = SetSlotToNode(srs, slot, targetNode.Id)
			if err != nil {
				return 0, err, ""
			}
			err = SetSlotStable(srs, slot)
			if err != nil {
				return 0, err, ""
			}
			trs := t.TargetReplicaSet()
			err = SetSlotToNode(trs, slot, targetNode.Id)
			if err != nil {
				return 0, err, ""
			}
			err = SetSlotStable(trs, slot)
			return 0, err, ""
		}
		return 0, err, ""
	}

	// 需要将Source分片的所有节点标记为MIGRATING，最大限度避免从地域的读造成的数据不一致
	for _, node := range rs.AllNodes() {
		err := redis.SetSlot(node.Addr(), slot, redis.SLOT_MIGRATING, targetNode.Id)
		if err != nil {
			if strings.HasPrefix(err.Error(), "ERR I'm not the owner of hash slot") {
				log.Warningf(t.TaskName(), "%s is not the owner of hash slot %d",
					sourceNode.Id, slot)
				srs := t.SourceReplicaSet()
				err = SetSlotStable(srs, slot)
				if err != nil {
					log.Warningf(t.TaskName(), "Failed to clean MIGRATING state of source server.")
					return 0, err, ""
				}
				trs := t.TargetReplicaSet()
				err = SetSlotStable(trs, slot)
				if err != nil {
					log.Warningf(t.TaskName(), "Failed to clean MIGRATING state of target server.")
					return 0, err, ""
				}
				return 0, fmt.Errorf("mig: %s is not the owner of hash slot %d", sourceNode.Id, slot), ""
			}
			return 0, err, ""
		}
	}

	/// 迁移的速度甚至迁移超时的配置可能都有不小问题，目前所有命令是短连接，且一次只迁移一个key

	// 一共迁移多少个key
	nkeys := 0
	app := meta.GetAppConfig()
	for {
		keys, err := redis.GetKeysInSlot(sourceNode.Addr(), slot, keysPer)
		if err != nil {
			return nkeys, err, ""
		}
		for _, key := range keys {
			_, err := redis.Migrate(sourceNode.Addr(), targetNode.Ip, targetNode.Port, key, app.MigrateTimeout)
			if err != nil {
				return nkeys, err, key
			}
			nkeys++
		}
		if len(keys) == 0 {
			// 迁移完成，需要等SourceSlaves同步(DEL)完成，即SourceSlaves节点中该slot内已无key
			slaveSyncDone := true
			srs := t.SourceReplicaSet()
			for _, node := range srs.AllNodes() {
				nkeys, err := redis.CountKeysInSlot(node.Addr(), slot)
				if err != nil {
					return nkeys, err, ""
				}
				if nkeys > 0 {
					slaveSyncDone = false
				}
			}
			if !slaveSyncDone {
				return nkeys, fmt.Errorf("mig: source nodes not all empty, will retry."), ""
			}
			// 设置slot归属到新节点，该操作自动清理IMPORTING和MIGRATING状态
			// 如果设置的是Source节点，设置slot归属时，Redis会确保该slot中已无剩余的key
			trs := t.TargetReplicaSet()
			// 优先设置从节点，保证当主的数据分布还未广播到从节点时主挂掉，slot信息也不会丢失
			for _, node := range trs.Slaves {
				if node.Fail {
					continue
				}
				err = redis.SetSlot(node.Addr(), slot, redis.SLOT_NODE, targetNode.Id)
				if err != nil {
					return nkeys, err, ""
				}
			}
			// 该操作增加Epoch并广播出去
			err = redis.SetSlot(trs.Master.Addr(), slot, redis.SLOT_NODE, targetNode.Id)
			if err != nil {
				return nkeys, err, ""
			}
			// 更新节点上slot的归属
			for _, rs := range t.cluster.ReplicaSets() {
				if rs.Master.IsStandbyMaster() {
					continue
				}
				err = SetSlotToNode(rs, slot, targetNode.Id)
				if err != nil {
					return nkeys, err, ""
				}
			}
			break
		}
	}

	return nkeys, nil, ""
}

func (t *MigrateTask) streamPub(careSpeed bool) {
	data := &streams.MigrateStateStreamData{
		SourceId:       t.SourceNode().Id,
		TargetId:       t.TargetNode().Id,
		State:          stateNames[t.CurrentState()],
		Ranges:         t.ranges,
		CurrRangeIndex: t.currRangeIndex,
		CurrSlot:       t.currSlot,
	}
	if careSpeed {
		now := time.Now()
		if now.Sub(t.lastPubTime) > 100*time.Millisecond {
			streams.MigrateStateStream.Pub(data)
			t.lastPubTime = now
		}
	} else {
		streams.MigrateStateStream.Pub(data)
	}
}

func (t *MigrateTask) Run() {
	prev_key := ""
	timeout_cnt := 0
	for i, r := range t.ranges {
		if r.Left < 0 {
			r.Left = 0
		}
		if r.Right > 16383 {
			r.Right = 16383
		}
		t.currRangeIndex = i
		t.currSlot = r.Left
		t.totalKeysInSlot = 0
		for t.currSlot <= r.Right {
			t.streamPub(true)

			// 尽量在迁移完一个完整Slot或遇到错误时，再进行状态的转换
			if t.CurrentState() == StateCancelling {
				t.SetState(StateCancelled)
				t.streamPub(false)
				return
			}

			// 暂停，sleep一会继续检查
			if t.CurrentState() == StatePausing {
				t.SetState(StatePaused)
			}
			if t.CurrentState() == StatePaused {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 正常运行
			app := meta.GetAppConfig()
			nkeys, err, key := t.migrateSlot(t.currSlot, app.MigrateKeysEachTime)
			t.totalKeysInSlot += nkeys
			// Check remains again
			seed := t.SourceNode()
			remains, err2 := redis.CountKeysInSlot(seed.Addr(), t.currSlot)
			if err2 != nil {
				remains = -1
			}
			if err != nil || remains > 0 {
				log.Warningf(t.TaskName(),
					"Migrate slot %d error, %d keys done, total %d keys, remains %d keys, %v",
					t.currSlot, nkeys, t.totalKeysInSlot, remains, err)
				if err != nil && strings.HasPrefix(err.Error(), "READONLY") {
					log.Warningf(t.TaskName(), "Migrating across slaves nodes. "+
						"Maybe a manual failover just happened, "+
						"if cluster marks down after this point, "+
						"we need recover it by ourself using cli commands.")
					t.SetState(StateCancelled)
					goto quit
				} else if err != nil && strings.HasPrefix(err.Error(), "CLUSTERDOWN") {
					log.Warningf(t.TaskName(), "The cluster is down, please check it yourself, migrating task cancelled.")
					t.SetState(StateCancelled)
					goto quit
				} else if err != nil && strings.HasPrefix(err.Error(), "IOERR") {
					log.Warningf(t.TaskName(), "Migrating key:%s timeout", key)
					if timeout_cnt > 10 {
						log.Warningf(t.TaskName(), "Migrating key:%s timeout too frequently, task cancelled")
						t.SetState(StateCancelled)
						goto quit
					}
					if prev_key == key {
						timeout_cnt++
					} else {
						timeout_cnt = 0
						prev_key = key
					}
				}
				time.Sleep(500 * time.Millisecond)
			} else {
				log.Infof(t.TaskName(), "Migrate slot %d done, %d keys done, total %d keys, remains %d keys",
					t.currSlot, nkeys, t.totalKeysInSlot, remains)
				t.currSlot++
				t.totalKeysInSlot = 0
			}
		}
	}
	t.currSlot--
	t.SetState(StateDone)
quit:
	t.streamPub(false)
}

func (t *MigrateTask) BackupReplicaSet() *topo.ReplicaSet {
	return t.backupReplicaSet
}

func (t *MigrateTask) SetBackupReplicaSet(rs *topo.ReplicaSet) {
	t.backupReplicaSet = rs
}

// 下面方法在MigrateManager中使用，需要原子操作

func (t *MigrateTask) CurrentState() int32 {
	return atomic.LoadInt32(&t.state)
}

func (t *MigrateTask) SetState(state int32) {
	atomic.StoreInt32(&t.state, state)
}

func (t *MigrateTask) ReplaceSourceReplicaSet(rs *topo.ReplicaSet) {
	t.source.Store(rs)
}

func (t *MigrateTask) ReplaceTargetReplicaSet(rs *topo.ReplicaSet) {
	t.target.Store(rs)
}

func (t *MigrateTask) SourceReplicaSet() *topo.ReplicaSet {
	return t.source.Load().(*topo.ReplicaSet)
}

func (t *MigrateTask) TargetReplicaSet() *topo.ReplicaSet {
	return t.target.Load().(*topo.ReplicaSet)
}

func (t *MigrateTask) SourceNode() *topo.Node {
	return t.SourceReplicaSet().Master
}

func (t *MigrateTask) TargetNode() *topo.Node {
	return t.TargetReplicaSet().Master
}
