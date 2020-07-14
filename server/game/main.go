package main

import (
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/service/balance"
	_ "github.com/greatwing/wing/server/game/model"
	"github.com/greatwing/wing/server/game/model/db"
	_ "github.com/greatwing/wing/server/gateway/backend"
	"time"
)

func main() {
	//初始化
	base.Init("game")

	//todo 连接数据库
	db.ConnectToMongoDB("mongodb://localhost:27017")
	db.ConnectToRedis("127.0.0.1:6379", "")

	//等待连接状态完成
	base.WaitReady()

	//继续保存之前没存完的数据
	db.ResumeSaveTask()

	////为mongodb的表添加索引
	//db.InitIndexes()

	//监听端口，并注册到etcd
	base.Accept(base.ServiceParameter{
		NetProcName: "svc.backend",
		ListenAddr:  ":0",
	})

	//上报负载
	balance.LoadReport(config.GetLocalSvcID(), func() int {
		return 1 //todo
	}, time.Second*60)

	//result, _ :=discovery.Default.CheckAndSet("test", "222")
	//log.Default.Infof("CheckAndSet [%s] [%s]", "222", result)

	base.StartLoop()
	base.Exit()
}
