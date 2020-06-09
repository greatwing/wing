package main

import (
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/proto"
	"reflect"
)

func main() {
	base.Init("GatewayServer")

	base.CreateCommnicateAcceptor(base.ServiceParameter{
		SvcName:     "game",
		NetProcName: "svc.backend",
		ListenAddr:  ":0",
	})

	base.ListenMsg(1, reflect.TypeOf(proto.UserInfo{}))

	base.StartLoop()
}
