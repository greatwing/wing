package db

import (
	"context"
	logger "github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/server/game/model/role"
	"github.com/mediocregopher/radix.v2/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type QueryRoleCallback func(r *role.Role, err error)

// 查询角色数据，单账号单角色，阻塞
func QueryRoleByUidSync(uid string) (*role.Role, error) {
	//查看cache
	roleBson, ok := getCachedRole(uid)
	if ok {
		//内存里面有，直接返回
		var r *role.Role
		err := bson.Unmarshal(roleBson.data, &r)
		if err == nil {
			r.RebuildItemMap()
			return r, nil
		}
	}

	//查看redis
	r := queryFromRedis(uid)
	if r != nil {
		return r, nil
	}

	//查看mongodb
	return queryFromMongoDB(uid)
}

// 查询角色数据，单账号单角色，非阻塞
func QueryRoleByUidAsync(uid string, callback QueryRoleCallback) {
	go func() {
		r, err := QueryRoleByUidSync(uid)
		if callback != nil {
			callback(r, err)
		}
	}()
}

func queryFromRedis(uid string) *role.Role {

	//t := time.Now()
	result := RedisConn.Operate(func(rawClient interface{}) interface{} {
		client := rawClient.(*redis.Client)
		data, err := client.Cmd("HGET", getRoleKey(uid), "d").Bytes()
		//data, err := client.Cmd("GET", getRoleKey(uid)).Bytes()
		if err != nil {
			if err != redis.ErrRespNil {
				logger.Errorf("query from redis err: %v", err)
			}
			return nil
		} else {
			var r *role.Role
			err = bson.Unmarshal(data, &r)
			if err == nil {
				r.RebuildItemMap()
				return r
			} else {
				return nil
			}
		}
	})

	if r, ok := result.(*role.Role); ok {
		//logger.Debugf("query role from redis cost: %v", time.Since(t))
		return r
	}
	return nil
}

func queryFromMongoDB(uid string) (*role.Role, error) {
	r := &role.Role{}
	//t := time.Now()

	//查询角色
	err := getCollection(CollectionRole).FindOne(context.Background(), bson.M{"uid": uid}).Decode(r)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//数据不存在
			logger.Infof("no role found by uid: %v", uid)
			return nil, nil
		} else {
			//其他错误
			logger.Error(err)
			return nil, err
		}
	}
	r.RebuildItemMap()

	////查询道具
	//cursor, err := getCollection(CollectionItem).Find(context.Background(), bson.M{"rid": r.ID})
	//if err == nil {
	//	var items []*proto.ItemData
	//	if err = cursor.All(context.Background(), &items); err == nil {
	//		r.AppendItems(items)
	//	} else {
	//		logger.Error(err)
	//	}
	//} else if err != mongo.ErrNoDocuments {
	//	logger.Error(err)
	//}

	//logger.Debugf("query role from mongodb cost: %v", time.Since(t))

	return r, nil
}
