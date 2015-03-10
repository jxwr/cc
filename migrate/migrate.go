package migrate

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/topo"
)

var (
	ErrActionTimeout = errors.New("migrate: action timeout")
)

type ActionType int
type Action struct {
	Type ActionType
	C    chan error
}

const (
	ACTION_PAUSE ActionType = iota
	ACTION_RESET
	ACTION_RESUME
	ACTION_CANCEL
)

type Range struct {
	Left  int
	Right int
}

type MigrateTask struct {
	// Node是创建迁移任务时的一个快照，它的信息可能会被更新
	// 这里仅使用Node的Ip,Port,Id的信息，其他信息不可用
	From   *topo.Node
	To     *topo.Node
	Ranges []Range

	currRangeIndex int // current range index
	currSlot       int // current slot
	actionChan     chan Action
}

func NewMigrateTask(fromNode, toNode *topo.Node, ranges []Range) *MigrateTask {
	t := &MigrateTask{
		From:       fromNode,
		To:         toNode,
		Ranges:     ranges,
		actionChan: make(chan Action),
	}
	return t
}

func (t *MigrateTask) migrateSlot(slot int, keysPer int) (int, error) {
	err := redis.SetSlot(t.From.Addr(), slot, redis.SLOT_MIGRATING, t.To.Id)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ERR I'm not the owner of hash slot") {
			log.Printf("%s is not the owner of hash slot %d\n", t.From.Id, slot)
			return 0, nil
		}
		return 0, err
	}
	err = redis.SetSlot(t.To.Addr(), slot, redis.SLOT_IMPORTING, t.From.Id)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ERR I'm already the owner of hash slot") {
			log.Printf("%s already the owner of hash slot %d\n", t.To.Id, slot)
			return 0, nil
		}
		return 0, err
	}

	/// 迁移的速度甚至迁移超时的配置可能都有不小问题，目前所有命令是短连接，且一次只迁移一个key

	// 一共迁移多少个key
	nkeys := 0
	for {
		// TODO: 流控，和迁移重试
		keys, err := redis.GetKeysInSlot(t.From.Addr(), slot, 100)
		if err != nil {
			return nkeys, err
		}
		for _, key := range keys {
			_, err := redis.Migrate(t.From.Addr(), t.To.Ip, t.To.Port, key, 15000)
			if err != nil {
				return nkeys, err
			}
			nkeys++
		}
		if len(keys) == 0 {
			// 迁移完成，设置slot归属到新节点，该操作自动清理IMPORTING和MIGRATING状态
			err = redis.SetSlot(t.From.Addr(), slot, redis.SLOT_NODE, t.To.Id)
			if err != nil {
				return nkeys, err
			}
			err = redis.SetSlot(t.To.Addr(), slot, redis.SLOT_NODE, t.To.Id)
			if err != nil {
				return nkeys, err
			}
			break
		}
	}

	return nkeys, nil
}

func (t *MigrateTask) Run() {
begin:
	pause := false

	for i, r := range t.Ranges {
		t.currRangeIndex = i
		t.currSlot = r.Left
		for t.currSlot < r.Right {
			// 每迁移完一整个slot或遇到错误处理一个动作
			// 如果状态停留在一次slot迁移内部，处理比较麻烦
			// 不过，只是尽量，还是有可能停在一个Slot内部
			select {
			case action := <-t.actionChan:
				switch action.Type {
				case ACTION_PAUSE:
					pause = true
					action.C <- nil
				case ACTION_RESUME:
					pause = false
					action.C <- nil
				case ACTION_CANCEL:
					action.C <- nil
					return
				case ACTION_RESET:
					action.C <- nil
					goto begin
				}
			default:
				// 无Action就继续迁移
			}

			if pause {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			nkeys, err := t.migrateSlot(t.currSlot, 100)
			if err != nil {
				log.Printf("Migrate slot %d error, %d keys have done, %v\n", t.currSlot, nkeys, err)
				time.Sleep(100 * time.Millisecond)
			} else {
				log.Printf("Migrate slot %d done, total %d keys\n", t.currSlot, nkeys)
				t.currSlot++
			}
		}
	}
}

func (t *MigrateTask) do(action ActionType, timeout time.Duration) error {
	c := make(chan error)
	timeoutChan := time.After(timeout)

	t.actionChan <- Action{action, c}
	select {
	case err := <-c:
		return err
	case <-timeoutChan:
		return ErrActionTimeout
	}
}

func (t *MigrateTask) Pause() error {
	err := t.do(ACTION_PAUSE, 1*time.Second)
	return err
}

func (t *MigrateTask) Resume() error {
	err := t.do(ACTION_RESUME, 1*time.Second)
	return err
}

func (t *MigrateTask) Cancel() error {
	err := t.do(ACTION_CANCEL, 2*time.Second)
	return err
}

func (t *MigrateTask) Reset() error {
	err := t.do(ACTION_RESET, 1*time.Second)
	return err
}
