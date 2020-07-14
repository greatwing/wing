package balance

import (
	"github.com/greatwing/wing/base/service/discovery"
	"github.com/greatwing/wing/base/timer"
	"strconv"
	"time"
)

type ReportFunc func() int

// 上报负载情况
func LoadReport(svcId string, fn ReportFunc, interval time.Duration) {
	if fn == nil {
		return
	}

	//立刻上报一次
	doReport(svcId, fn)

	//定时上报
	timer.Schedule(func(dt time.Duration) {
		doReport(svcId, fn)
	}, interval, 0, 0)
}

func doReport(svcId string, fn ReportFunc) {
	connections := fn()
	go discovery.Default.SetValue(svcId, strconv.Itoa(connections), discovery.WithBalanceKey(), discovery.WithLease())
}
