package mongodb

import (
	"github.com/davyxu/cellnet"
	"go.mongodb.org/mongo-driver/mongo"
)

type Connector interface {
	cellnet.GenericPeer

	// 修改默认的数据库名
	SetDefaultDBName(dbName string)

	Database() *mongo.Database
	Collection(name string) *mongo.Collection
}
