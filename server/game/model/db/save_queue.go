package db

import (
	"context"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/config"
	logger "github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/msg"
	"github.com/greatwing/wing/base/timer"
	"github.com/greatwing/wing/server/game/model/role"
	"github.com/mediocregopher/radix.v2/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"sync"
	"time"
)

type roleBsonData struct {
	rid  primitive.ObjectID
	data []byte
}

var (
	roleCacheMap   = make(map[string]roleBsonData)
	roleCacheGuard sync.RWMutex
	roleTimer      = timer.New(nil)
	roleRedisTime  sync.Map
)

//type itemExt struct {
//	*proto.ItemData `bson:"rid,omitempty,inline"`
//	RoleId          primitive.ObjectID `bson:"rid,omitempty"`
//}

const (
	saveDbDuration = time.Second * 1 //角色数据多长时间保存一次到mongodb
	redisTTL       = 24 * 60 * 60    //数据保存到monogodb后，redis中多数据多长时间超时
)

// 保存角色数据
func SaveRole(r *role.Role) {
	//修改更新时间
	r.UpdateTime = time.Now().Unix()

	//串行化
	bsonData, err := bson.Marshal(r)
	if err != nil {
		logger.Error(err)
		return
	}

	//过一段时间后存入MongoDB
	saveRoleLater(r.Uid, r.ID, bsonData, saveDbDuration)

	//立刻写入redis
	go func() {
		t := time.Now()
		ts := t.UnixNano()
		roleRedisTime.Store(r.Uid, ts)
		RedisConn.Operate(func(rawClient interface{}) interface{} {
			client := rawClient.(*redis.Client)

			//client.Cmd("MULTI")

			//记录待保存到mongodb的任务
			client.Cmd("SADD", getSavingTaskKey(), r.Uid)

			//保存角色数据，并记录时间戳
			rsp := client.Cmd("HMSET", getRoleKey(r.Uid), "d", bsonData, "ts", ts,
				"svc", config.GetLocalSvcID())
			//rsp := client.Cmd("EXEC")

			if rsp.Err != nil {
				logger.Errorf("save role to redis err: %v", rsp.Err)
			} else {
				logger.Debugf("save to redis uid=%v, cost: %v", r.Uid, time.Since(t))
			}

			return nil
		})
	}()
}

func IsRoleCached(uid string) bool {
	roleCacheGuard.Lock()
	defer roleCacheGuard.Unlock()

	_, ok := roleCacheMap[uid]
	return ok
}

func cacheRole(uid string, bsonData roleBsonData) {
	roleCacheGuard.Lock()
	defer roleCacheGuard.Unlock()

	roleCacheMap[uid] = bsonData
}

func getCachedRole(uid string) (roleBsonData, bool) {
	roleCacheGuard.RLock()
	defer roleCacheGuard.RUnlock()

	data, ok := roleCacheMap[uid]
	return data, ok
}

func removeCachedRole(uid string) (roleBsonData, bool) {
	roleCacheGuard.Lock()
	defer roleCacheGuard.Unlock()

	if data, ok := roleCacheMap[uid]; ok {
		delete(roleCacheMap, uid)
		return data, true
	}

	return roleBsonData{}, false
}

func saveRoleLater(uid string, rid primitive.ObjectID, data []byte, delay time.Duration) {
	noSchedule := !IsRoleCached(uid)
	cacheRole(uid, roleBsonData{
		rid:  rid,
		data: data,
	})

	if noSchedule {
		//定时保存
		roleTimer.Run(func(dt time.Duration) {
			saveToMongoDb(uid)
		}, delay, 1)
	}
}

func saveToMongoDb(uid string) {

	//先从内存里移除
	roleBson, ok := removeCachedRole(uid)
	if !ok {
		return
	}

	//t := time.Now()
	var ts int64
	if value, ok := roleRedisTime.Load(uid); ok {
		ts = value.(int64)
	}

	//存角色数据
	_, err := getCollection(CollectionRole).ReplaceOne(context.Background(),
		bson.M{"_id": roleBson.rid},
		roleBson.data,
		options.Replace().SetUpsert(true))
	if err != nil {
		//保存失败
		logger.Errorf("saveToMongoDb uid=%v, rid=%v, err: %v", uid, roleBson.rid, err)
		return
	}

	////保存道具
	//updatedItems := r.GetUpdatedItem()
	//for _, item := range updatedItems {
	//	newItem := itemExt{
	//		ItemData: item,
	//		RoleId:   r.ID,
	//	}
	//	_, err = getCollection(CollectionItem).ReplaceOne(context.Background(),
	//		bson.M{"_id": item.UUID},
	//		newItem,
	//		options.Replace().SetUpsert(true))
	//	if err != nil {
	//		logger.Error(err)
	//	}
	//}

	//logger.Debugf("saveToMongoDb uid=%v, cost: %v", uid, time.Since(t))

	//修改redis的ttl
	if !IsRoleCached(uid) {
		RedisConn.Operate(func(rawClient interface{}) interface{} {
			client := rawClient.(*redis.Client)
			key := getRoleKey(uid)

			client.Cmd("WATCH", key)
			isWatching := true
			defer func() {
				if isWatching {
					client.Cmd("UNWATCH")
				}
			}()

			savedTS, err := client.Cmd("HGET", key, "ts").Int64()
			if err != nil {
				logger.Error(err)
				return nil
			}
			if ts != savedTS {
				//和记录的时间戳不同，不设置超时
				return nil
			}

			client.Cmd("MULTI")
			client.Cmd("EXPIRE", key, redisTTL)
			//client.Cmd("HDEL", key, "svc")
			client.Cmd("SREM", getSavingTaskKey(), uid)
			rsp := client.Cmd("EXEC")
			isWatching = false

			if rsp.Err != nil {
				logger.Error(rsp.Err)
			} else if rsp.IsType(redis.Nil) {
				// EXEC 返回nil-reply来表示事务已经失败，监视的键在 EXEC 执行之前被修改了
			}

			return nil
		})
	}

	//保存成功，判断要不要删除user location
	base.RunInLogicGoroutine(func() {
		if !IsRoleCached(uid) {
			msg.FireCustom("DelUserLocation", uid)
		}
	})
}

// 把已记录在redis里，但是还没保存到mongodb的数据保存一下
func ResumeSaveTask() {
	RedisConn.Operate(func(rawClient interface{}) interface{} {
		client := rawClient.(*redis.Client)
		tasks, err := client.Cmd("SMEMBERS", getSavingTaskKey()).List()
		if err != nil {
			return nil
		}
		for _, uid := range tasks {
			logger.Infof("prepare to save %v from redis to mongdb", uid)
			data, err := client.Cmd("HMGET", getRoleKey(uid), "d", "ts").ListBytes()
			if err != nil || len(data) < 2 || data[0] == nil {
				continue
			}

			var r *role.Role
			err = bson.Unmarshal(data[0], &r)
			if err == nil {
				if data[1] != nil {
					ts, err := strconv.ParseInt(string(data[1]), 10, 64)
					if err == nil {
						roleRedisTime.Store(uid, ts)
					}
				}
				saveRoleLater(uid, r.ID, data[0], saveDbDuration)
			}
		}
		return nil
	})
}

func getSavingTaskKey() string {
	return "/savetask" + config.GetLocalSvcID()
}
