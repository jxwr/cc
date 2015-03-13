package migrate

import (
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/topo"
)

const (
	StateRunning int32 = iota
	StatePausing
	StatePaused
	StateCancelling
	StateCancelled
	StateNodeFailure
)

type MigrateTask struct {
	// Node是创建迁移任务时的一个快照，它的信息可能会被更新
	// 这里仅使用Node的Ip,Port,Id的信息，其他信息不可用
	Ranges []topo.Range

	source         atomic.Value
	target         atomic.Value
	currRangeIndex int // current range index
	currSlot       int // current slot
	state          int32
}

func NewMigrateTask(sourceRS, targetRS *topo.ReplicaSet, ranges []topo.Range) *MigrateTask {
	t := &MigrateTask{
		Ranges: ranges,
		state:  StateRunning,
	}
	t.ReplaceSourceReplicaSet(sourceRS)
	t.ReplaceTargetReplicaSet(targetRS)
	return t
}

func (t *MigrateTask) TaskName() string {
	return fmt.Sprintf("Mig[%s->%s]", t.SourceNode().Id, t.TargetNode().Id)
}

func (t *MigrateTask) migrateSlot(slot int, keysPer int) (int, error) {
	rs := t.SourceReplicaSet()
	sourceNode := t.SourceNode()
	targetNode := t.TargetNode()

	// 需要将Source分片的所有节点标记为MIGRATING，最大限度避免从地域的读造成的数据不一致
	// 这样操作降低问题的严重性，但由于是异步同步数据，读取到旧数据还是有小概率发生
	for _, node := range rs.AllNodes() {
		err := redis.SetSlot(node.Addr(), slot, redis.SLOT_MIGRATING, targetNode.Id)
		if err != nil {
			if strings.HasPrefix(err.Error(), "ERR I'm not the owner of hash slot") {
				log.Printf("%s %s is not the owner of hash slot %d\n",
					t.TaskName(), sourceNode.Id, slot)
				return 0, nil
			}
			return 0, err
		}
	}

	err := redis.SetSlot(targetNode.Addr(), slot, redis.SLOT_IMPORTING, sourceNode.Id)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ERR I'm already the owner of hash slot") {
			log.Printf("%s %s already the owner of hash slot %d\n",
				t.TaskName(), targetNode.Id, slot)
			return 0, nil
		}
		return 0, err
	}

	/// 迁移的速度甚至迁移超时的配置可能都有不小问题，目前所有命令是短连接，且一次只迁移一个key

	// 一共迁移多少个key
	nkeys := 0
	for {
		// TODO: 流控，和迁移重试
		keys, err := redis.GetKeysInSlot(sourceNode.Addr(), slot, 100)
		if err != nil {
			return nkeys, err
		}
		for _, key := range keys {
			_, err := redis.Migrate(sourceNode.Addr(), targetNode.Ip, targetNode.Port, key, 15000)
			if err != nil {
				return nkeys, err
			}
			nkeys++
		}
		if len(keys) == 0 {
			// 迁移完成，设置slot归属到新节点，该操作自动清理IMPORTING和MIGRATING状态
			// 如果设置的是Source节点，设置slot归属时，Redis会确保该slot中已无剩余的key
			// 即便slot不属于这个节点，该操作也会成功
			rs := t.SourceReplicaSet()
			// 首先更新Master节点，主的slot状态会被Controller看见
			err = redis.SetSlot(rs.Master().Addr(), slot, redis.SLOT_NODE, targetNode.Id)
			if err != nil {
				return nkeys, err
			}
			// 清理从节点的MIGRATING状态
			for _, node := range rs.Slaves() {
				err = redis.SetSlot(node.Addr(), slot, redis.SLOT_NODE, targetNode.Id)
				if err != nil {
					return nkeys, err
				}
			}
			// 更新slot在目标节点上的归属，该操作增加Epoch，进而广播出去
			err = redis.SetSlot(targetNode.Addr(), slot, redis.SLOT_NODE, targetNode.Id)
			if err != nil {
				return nkeys, err
			}
			break
		}
	}

	return nkeys, nil
}

func (t *MigrateTask) Run() {
	for i, r := range t.Ranges {
		t.currRangeIndex = i
		t.currSlot = r.Left
		for t.currSlot < r.Right {
			// 尽量在迁移完一个完整Slot或遇到错误时，再进行状态的转换
			// 只是尽量而已，还是有可能停在一个Slot内部

			if t.CurrentState() == StateCancelling {
				t.SetState(StateCancelled)
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

			// Source节点或Target节点挂了，一般这时会遇到错误，不必
			// 多增加一个ING状态
			if t.CurrentState() == StateNodeFailure {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 正常运行
			nkeys, err := t.migrateSlot(t.currSlot, 100)
			if err != nil {
				log.Printf("%s Migrate slot %d error, %d keys have done, %v\n",
					t.TaskName(), t.currSlot, nkeys, err)
				time.Sleep(100 * time.Millisecond)
			} else {
				log.Printf("%s Migrate slot %d done, total %d keys\n",
					t.TaskName(), t.currSlot, nkeys)
				t.currSlot++
			}
		}
	}
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
	return t.SourceReplicaSet().Master()
}

func (t *MigrateTask) TargetNode() *topo.Node {
	return t.TargetReplicaSet().Master()
}
