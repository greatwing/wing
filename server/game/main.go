package main

import (
	"github.com/greatwing/wing/base"
	_ "github.com/greatwing/wing/server/game/msghandler"
	_ "github.com/greatwing/wing/server/gateway/backend"
)

func main() {
	//初始化
	base.Init("game")

	//监听端口，并注册到etcd
	base.Accept(base.ServiceParameter{
		NetProcName: "svc.backend",
		ListenAddr:  ":0",
	})

	base.CheckReady(nil)

	base.StartLoop()
	base.Exit()
}
