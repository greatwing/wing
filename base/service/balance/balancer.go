package balance

import (
	"errors"
	"github.com/greatwing/wing/base/service/discovery"
	"sort"
	"strconv"
	"sync"
)

type LoadBalancer interface {
	LeastConnections(count int) ([]string, error)
	SetLoad(svcID string, value string)
	Remove(svcID string)
}

type balancer struct {
	loadMap      map[string]*loadInfo
	loadSlice    []*loadInfo
	isSorted     bool
	loadMapGuard sync.RWMutex
}

type loadInfo struct {
	svcID string
	load  int
}

func (b *balancer) SetLoad(svcID string, value string) {
	load, err := strconv.Atoi(value)
	if err != nil {
		return
	}

	b.loadMapGuard.Lock()
	defer b.loadMapGuard.Unlock()

	if info, ok := b.loadMap[svcID]; ok {
		info.load = load
	} else {
		info = &loadInfo{svcID: svcID, load: load}
		b.loadMap[svcID] = info
		b.loadSlice = append(b.loadSlice, info)
	}

	//需要重新排序
	b.isSorted = false
}

func (b *balancer) AddLoad(svcID string, add int) {
	b.loadMapGuard.Lock()
	defer b.loadMapGuard.Unlock()

	if info, ok := b.loadMap[svcID]; ok {
		info.load += add
	} else {
		info = &loadInfo{svcID: svcID, load: add}
		b.loadMap[svcID] = info
		b.loadSlice = append(b.loadSlice, info)
	}

	//需要重新排序
	b.isSorted = false
}

func (b *balancer) Remove(svcID string) {
	b.loadMapGuard.Lock()
	defer b.loadMapGuard.Unlock()

	//删除操作不需要重新排序
	delete(b.loadMap, svcID)
	for index, info := range b.loadSlice {
		if info.svcID == svcID {
			b.loadSlice = append(b.loadSlice[:index], b.loadSlice[index+1:]...)
			break
		}
	}
}

// 获取连接数最少的svc
func (b *balancer) LeastConnections(count int) ([]string, error) {
	b.loadMapGuard.RLock()
	defer b.loadMapGuard.RUnlock()

	if len(b.loadSlice) == 0 {
		return nil, errors.New("no svc at all")
	}

	if !b.isSorted {
		sort.Slice(b.loadSlice, func(i, j int) bool {
			return b.loadSlice[i].load < b.loadSlice[j].load
		})

		b.isSorted = true
	}

	result := make([]string, 0, 5)

	for index, info := range b.loadSlice {
		if index >= count {
			break
		}
		result = append(result, info.svcID)
	}

	return result, nil
}

func New(svcName string) LoadBalancer {
	b := &balancer{
		loadMap: make(map[string]*loadInfo),
	}

	discovery.Default.WatchKey(svcName, func(op discovery.OperateType, key string, value string) {
		svcID := discovery.GetSvcIDByBalanceKey(key)
		if svcID == "" {
			return
		}

		switch op {
		case discovery.PUT:
			b.SetLoad(svcID, value)
		case discovery.DELETE:
			b.Remove(svcID)
		}
	}, discovery.WithBalanceKey(), discovery.WithPrefix())

	return b
}
