package redis

//var LocationKeyPrefix = "/location"
//
//type Location struct {
//	SvcID string `json:"svc_id"`
//}
//
//func GetLocationKey(id string) string {
//	if !strings.HasPrefix(id, "/") {
//		id = "/" + id
//	}
//	return LocationKeyPrefix + id
//}
//
//func IsLocationKey(rawkey string) bool {
//
//	return strings.HasPrefix(rawkey, LocationKeyPrefix)
//}
//
//// 获取位置，阻塞
//func GetLocation(id string) *Location {
//
//	key := GetLocationKey(id)
//
//	locData := Operator.Operate(func(rawClient interface{}) interface{} {
//		client := rawClient.(*redis.Client)
//		data, err := client.Cmd("GET", key).Bytes()
//		if err != nil {
//			logger.Error(err)
//			return nil
//		}
//
//		loc := Location{}
//		err = json.Unmarshal(data, &loc)
//		if err != nil {
//			logger.Error(err)
//			return nil
//		}
//
//		return &loc
//	})
//
//	if loc, ok := locData.(*Location); ok {
//		return loc
//	} else {
//		return nil
//	}
//}
//
//// 上报位置，阻塞
//func ReportLocation(id string) error {
//	key := GetLocationKey(id)
//	loc := Location{
//		SvcID: config.GetLocalSvcID(),
//	}
//
//	data, err := json.Marshal(loc)
//	if err != nil {
//		logger.Error(err)
//		return err
//	}
//
//	Operator.Operate(func(rawClient interface{}) interface{} {
//		client := rawClient.(*redis.Client)
//		client.Cmd("SET", key, string(data))
//		return nil
//	})
//
//	return nil
//}
//
//func TransactionSetLocation(id string, svcId string) string {
//	key := GetLocationKey(id)
//	result := Operator.Operate(func(rawClient interface{}) interface{} {
//		client := rawClient.(*redis.Client)
//
//		for {
//			client.Cmd("WATCH", key)
//			data, err := client.Cmd("GET", key).Str()
//			if err != nil && err != redis.ErrRespNil {
//				logger.Error(err)
//				return nil
//			}
//			logger.Infof("TransactionSetLocation GET: %s", data)
//
//			client.Cmd("MULTI")
//			if data == "" {
//				data = svcId
//				client.Cmd("SET", key, data)
//			} else {
//				client.Cmd("PERSIST", key)
//			}
//
//			rsp := client.Cmd("EXEC")
//			if rsp.Err != nil {
//				logger.Error(rsp.Err)
//				return nil
//			}
//
//			if rsp.IsType(redis.Nil) {
//				// EXEC 返回nil-reply来表示事务已经失败，监视的键在 EXEC 执行之前被修改了
//				logger.Infof("redis watch key modified, retry")
//				continue
//			}
//
//			return data
//		}
//	})
//
//	if str, ok := result.(string); ok {
//		return str
//	} else {
//		return ""
//	}
//}
