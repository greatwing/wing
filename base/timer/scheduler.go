package timer

import (
	"github.com/greatwing/wing/base"
	"time"
)

type Stopper interface {
	Stop()
}

type ScheduleFunc func(dt time.Duration)

// 定时计划
//  fn 			回调函数
//  interval 	定时间隔
//  repeat 		重复次数，0代表无限次数
//  delay		延迟启动定时计划
// 返回值：
//  Stopper		可以调用Stop提前结束定时计划
func Schedule(fn ScheduleFunc, interval time.Duration, repeat int, delay time.Duration) Stopper {
	t := New(base.Queue)
	if delay > 0 {
		//延迟
		t.Run(func(dt time.Duration) {
			t.Run(fn, interval, repeat)
		}, delay, 1)
	} else {
		t.Run(fn, interval, repeat)
	}
	return t
}

// 定时计划执行一次
//	fn 			回调函数
//  interval 	定时间隔
//  delay		延迟启动定时计划
// 返回值：
//  Stopper		可以调用Stop提前结束定时计划
func ScheduleOnce(fn ScheduleFunc, interval time.Duration, delay time.Duration) Stopper {
	return Schedule(fn, interval, 1, delay)
}
