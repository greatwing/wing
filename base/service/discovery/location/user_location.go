package location

import (
	"context"
	"fmt"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service"
	"github.com/greatwing/wing/base/service/balance"
	"github.com/greatwing/wing/base/service/discovery"
	"sync"
)

var (
	balancer    balance.LoadBalancer
	deletingMap sync.Map
)

// 选个一个game节点，返回service id
//  uid 用户id
func GetUserLocation(uid string) string {

	svcID, err := discovery.Default.GetValue("/user/"+uid, discovery.WithLocationKey())
	if err == nil && svcID != "" {
		ses := service.GetRemoteService(svcID)
		if ses != nil {
			return svcID
		} else {
			//连接不上,删除这个svcId
			_ = discovery.Default.DelValue(svcID)
		}
	}

	if balancer == nil {
		balancer = balance.New("game")
	}
	for {
		//负载均衡给一个game
		ret, err := balancer.LeastConnections(1)
		if err == nil && len(ret) > 0 {
			svcID = ret[0]

			ses := service.GetRemoteService(svcID)
			if ses != nil {
				return svcID
			} else {
				//这个svc连不上，从负载均衡里删除它
				balancer.Remove(svcID)
			}

		} else {
			break
		}
	}
	return ""
}

func SetUserLocation(uid, svcId string) error {
	if value, ok := deletingMap.Load(uid); ok {
		if cancel, ok := value.(context.CancelFunc); ok {
			//取消正在进行的delete操作
			cancel()
		}
	}

	result, err := discovery.Default.CheckAndSet("/user/"+uid, svcId,
		discovery.WithLocationKey(),
		discovery.WithLease())
	if err != nil {
		logger.Error(err)
		return err
	}

	if result != svcId {
		//已经被其他节点抢注
		return fmt.Errorf("uid[%s] has been set by [%s]", uid, result)
	}
	return nil
}

func DelUserLocation(uid, svcId string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deletingMap.Store(uid, cancel)
	err := discovery.Default.IfDel(ctx, "/user/"+uid, svcId, discovery.WithLocationKey())
	deletingMap.Delete(uid)

	return err
}
