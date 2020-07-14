package db

import (
	"context"
	logger "github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/peer/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	MongoConn   mongodb.Connector
	collections = make(map[string]*mongo.Collection)
)

const (
	CollectionRole = "role"
	CollectionItem = "item"
)

func ConnectToMongoDB(connectStr string) {
	MongoConn = mongodb.Connect(connectStr)
}

func InitIndexes() {
	view := getCollection(CollectionItem).Indexes()
	cursor, err := view.List(context.Background())
	if err != nil {
		logger.Fatal(err)
	}

	for cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			logger.Fatal(err)
		}
		logger.Infof("mongodb item indexes: %v", result)
	}
	if err := cursor.Err(); err != nil {
		logger.Fatal(err)
	}
	cursor.Close(context.Background())
}

func getCollection(name string) *mongo.Collection {
	if coll, ok := collections[name]; ok {
		return coll
	} else {
		coll = MongoConn.Collection(name)
		collections[name] = coll
		return coll
	}
}
