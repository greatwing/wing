package main

import (
	"fmt"
	_ "github.com/davyxu/cellnet/peer/gorillaws"
	_ "github.com/davyxu/cellnet/peer/redix"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/service"
	_ "github.com/greatwing/wing/server/gateway/backend"
	"github.com/greatwing/wing/server/gateway/frontend"
	"github.com/greatwing/wing/server/gateway/route"
)

func main() {
	//初始化
	base.Init("gateway")

	//连接game服务
	mp := base.Connect(base.ServiceParameter{
		SvcName:     "game",
		NetProcName: "gateway.backend",
	}, service.FilterMatchRule([]string{"/game/*"}))

	mp.SkipReadyCheck() //不检测连接game的ready状态

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
		//todo
	}

	//添加路由规则
	route.AddRules(10001, 65535, "game")

	//连接redis
	base.ConnectToRedis("127.0.0.1:6379")

	//检测peer的ready状态
	base.CheckReady(nil)

	base.StartLoop()
	base.Exit()
}
