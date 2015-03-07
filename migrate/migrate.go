package migrate

import (
	"errors"
	"fmt"
	"time"
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
	From   string
	To     string
	Ranges []Range

	currRangeIndex int // current range index
	currSlot       int // current slot
	actionChan     chan Action
}

func NewMigrateTask(fromId, toId string, ranges []Range) *MigrateTask {
	t := &MigrateTask{
		From:       fromId,
		To:         toId,
		Ranges:     ranges,
		actionChan: make(chan Action),
	}
	return t
}

func (t *MigrateTask) migrateSlot(slot int) error {
	return nil
}

func (t *MigrateTask) Run() {
begin:
	pause := false

	for i, r := range t.Ranges {
		t.currRangeIndex = i
		t.currSlot = r.Left
		for t.currSlot < r.Right {
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
			}

			time.Sleep(1000 * time.Millisecond)

			if pause {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			err := t.migrateSlot(t.currSlot)
			if err != nil {
				fmt.Printf("migrate %d error\n", t.currSlot)
				time.Sleep(100 * time.Millisecond)
			} else {
				fmt.Printf("migrate %d success\n", t.currSlot)
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
