package mongodb

import (
	"context"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"sync"
	"time"
)

type mongodbConnector struct {
	peer.CorePeerProperty
	peer.CoreContextSet

	client      *mongo.Client
	clientGuard sync.RWMutex

	dbName string
	db     *mongo.Database
}

func (m *mongodbConnector) TypeName() string {
	return "mongodb.Connector"
}

func (m *mongodbConnector) Start() cellnet.Peer {

	go m.tryConnect()

	return m
}

func (m *mongodbConnector) Client() *mongo.Client {
	m.clientGuard.RLock()
	defer m.clientGuard.RUnlock()

	return m.client
}

func (m *mongodbConnector) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.Client().Disconnect(ctx); err != nil {
		logger.Error(err)
	}
}

func (m *mongodbConnector) IsReady() bool {
	return m.Client() != nil
}

func (m *mongodbConnector) tryConnect() {
	for {
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(m.Address()))
		if err != nil {
			logger.Error(err)
			continue
		}

		err = client.Ping(context.Background(), readpref.Primary())
		if err != nil {
			logger.Error(err)
			client.Disconnect(context.Background())
			continue
		}

		m.clientGuard.Lock()
		m.client = client
		m.clientGuard.Unlock()

		break
	}
}

func (m *mongodbConnector) SetDefaultDBName(dbName string) {
	m.clientGuard.RLock()
	defer m.clientGuard.RUnlock()

	m.dbName = dbName
	m.db = nil
}

func (m *mongodbConnector) Database() *mongo.Database {
	m.clientGuard.RLock()
	defer m.clientGuard.RUnlock()

	if m.db == nil {
		if m.dbName == "" {
			m.dbName = "db_" + config.GetSvcGroup()
		}
		m.db = m.client.Database(m.dbName)
	}
	return m.db
}

func (m *mongodbConnector) Collection(name string) *mongo.Collection {
	return m.Database().Collection(name)
}

func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {

		self := &mongodbConnector{}
		//self.CoreRedisParameter.Init()

		return self
	})
}
