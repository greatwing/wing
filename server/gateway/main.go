package main

import (
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/golog"
	"github.com/greatwing/wing/base"
	_ "github.com/greatwing/wing/server/gateway/backend"
)

var log = golog.New("gateway")

func main() {
	base.Init("gate")

	// 要连接的服务列表
	base.CreateCommnicateConnector(base.ServiceParameter{
		SvcName:      "game",
		MaxConnCount: -1,
		NetProcName:  "agent.backend",
	})

	//base.CreateCommnicateAcceptor(base.ServiceParameter{
	//	SvcName:     "gate",
	//	NetProcName: "svc.backend",
	//	ListenAddr:  ":18801",
	//})

	//base.ListenMsg(1, func(ev cellnet.Event) {
	//	if msg, ok := ev.Message().(*proto.UserInfo); ok {
	//		log.Infof("msg:%v, len:%v, cnt:%d", msg.Message, msg.Length, msg.Cnt)
	//		ev.Session().Send(&proto.UserInfo{
	//			Message: "I am GatewayServer",
	//			Length:  10,
	//			Cnt:     20,
	//		})
	//	} else {
	//		log.Errorf("decode error msgId:%d", 1)
	//	}
	//})

	//todo
	base.StartLoop()

	base.Exit()
}
