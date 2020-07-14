package main

import (
	"fmt"
	_ "github.com/davyxu/cellnet/peer/gorillaws"
	_ "github.com/davyxu/cellnet/peer/redix"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service"
	"github.com/greatwing/wing/base/service/balance"
	_ "github.com/greatwing/wing/server/gateway/backend"
	"github.com/greatwing/wing/server/gateway/frontend"
	"github.com/greatwing/wing/server/gateway/route"
	"time"
)

func main() {
	//初始化
	base.Init("gateway")

	//连接game服务
	mp := base.Connect(base.ServiceParameter{
		SvcName:     "game",
		NetProcName: "gateway.backend",
	}, service.FilterMatchRule([]string{"/game/*"}))
	mp.SetSkipCheck(true) //不检测连接game的ready状态

	//添加路由规则
	route.AddRules(10001, 65535, "game")

	//监听客户端连接的端口
	frontendAddr := fmt.Sprintf("%s:%s", config.GetWANIP(), config.GetPortsRange())
	switch config.GetProtocol() {
	case "tcp":
		frontend.Accept(base.ServiceParameter{
			SvcName:     "gateway",
			ListenAddr:  frontendAddr,
			NetPeerType: "tcp.Acceptor",
			NetProcName: "tcp.frontend",
		})
	case "ws":
		//websocket
		frontend.Accept(base.ServiceParameter{
			SvcName:     "gateway_ws",
			ListenAddr:  frontendAddr,
			NetPeerType: "gorillaws.Acceptor",
			NetProcName: "ws.frontend",
		})
	case "wss":
		//todo wss支持
	}

	//前端心跳检测
	frontend.StartHeartCheck()

	//上报负载
	balance.LoadReport(config.GetLocalSvcID(), func() int {
		load := frontend.FrontendSessionManager.Count()
		logger.Debugf("load report: %v", load)
		return load
	}, time.Second*5)

	base.StartLoop()
	base.Exit()
}
