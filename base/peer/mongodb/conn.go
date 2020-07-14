package mongodb

import (
	"github.com/davyxu/cellnet/peer"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/service"
)

// 连接MongoDB
//	connectStr 连接串，格式为 mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
func Connect(connectStr string) Connector {
	p := peer.NewGenericPeer("mongodb.Connector", "mongodb", connectStr, base.Queue)

	p.Start() //开始连接

	service.AddLocalService(p)

	return p.(Connector)
}
