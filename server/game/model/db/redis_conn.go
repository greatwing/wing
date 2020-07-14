package db

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/redix"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/service"
)

var (
	RedisConn cellnet.RedisPoolOperator
)

func ConnectToRedis(address, password string) {
	p := peer.NewGenericPeer("redix.Connector", "redis", address, base.Queue)

	conn := p.(cellnet.RedisConnector)
	conn.SetPassword(password) //密码

	p.Start() //开始连接

	service.AddLocalService(p)

	RedisConn = p.(cellnet.RedisPoolOperator)
}

func getRoleKey(uid string) string {
	return "/role/" + uid
}
