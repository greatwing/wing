package frontend

import (
	logger "github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/timer"
	"github.com/greatwing/wing/server/gateway/model"
	"time"
)

const HeartBeatDuration = time.Second * 120

// 心跳检测
func StartHeartCheck() {
	t := timer.New(nil)
	t.Run(func(dt time.Duration) {
		now := time.Now()
		VisitClient(func(client *model.Client) bool {
			if now.Sub(client.LastPingTime) > HeartBeatDuration {
				//超时踢下线
				logger.Infof("Close client due to heatbeat time out, id: %d", client.ClientSession.ID())
				client.ClientSession.Close()
			}
			return true
		})
	}, HeartBeatDuration, 0)
}
