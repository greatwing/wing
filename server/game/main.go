package main

import (
	"github.com/davyxu/golog"
	"github.com/greatwing/wing/base"
)

var log = golog.New("logic")

func main() {
	base.Init("game")

	base.CreateCommnicateAcceptor(base.ServiceParameter{
		SvcName:     "game",
		NetProcName: "svc.backend",
		ListenAddr:  ":0",
	})

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
