package timer

import (
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/log"
	"sync"
	"time"
)

type Runner interface {
	Run(fn ScheduleFunc, d time.Duration, repeat int)
	IsRunning() bool
	Stop()
}

type timer struct {
	stopCh     chan struct{}
	innerTimer *time.Timer
	guard      sync.RWMutex
	que        cellnet.EventQueue
}

func New(que cellnet.EventQueue) Runner {
	return &timer{
		que:    que,
		stopCh: make(chan struct{}),
	}
}

func (t *timer) Run(fn ScheduleFunc, d time.Duration, repeat int) {
	if repeat < 0 {
		logger.Panic("repeat count below zero")
	}

	//重置timer状态
	t.guard.Lock()
	if t.innerTimer == nil {
		t.innerTimer = time.NewTimer(d)
	} else {
		t.innerTimer.Reset(d)
	}
	if t.stopCh == nil {
		t.stopCh = make(chan struct{})
	}
	ch := t.stopCh
	t.guard.Unlock()

	beginTime := time.Now()

	go func() {
		count := 0
		for {
			select {
			case nowTime := <-t.innerTimer.C:

				count++
				if repeat == 0 || count < repeat {
					t.innerTimer.Reset(d)
				} else {
					t.Stop()
				}

				if fn != nil {
					//log.logger.Debugf("timer %d", count)
					cellnet.QueuedCall(t.que, func() {
						fn(nowTime.Sub(beginTime))
					})
				}

			case <-ch:
				return
			}

		}
	}()
}

// 查看定时器是否还在运行
func (t *timer) IsRunning() bool {
	t.guard.RLock()
	defer t.guard.RUnlock()

	return t.stopCh != nil
}

// 停止定时器
func (t *timer) Stop() {
	t.guard.Lock()
	defer t.guard.Unlock()

	if t.innerTimer != nil {
		t.innerTimer.Stop()
	}

	if t.stopCh != nil {
		close(t.stopCh)
		t.stopCh = nil
	}
}
