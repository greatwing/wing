package main

import (
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/davyxu/cellnet/peer/http"
	_ "github.com/davyxu/cellnet/peer/redix"
	_ "github.com/davyxu/cellnet/proc/http"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service/balance"
	"github.com/greatwing/wing/base/service/discovery"
	"github.com/greatwing/wing/server/login/responce"
	"github.com/greatwing/wing/server/login/token"
	"net/http"
	"time"
)

//var redisOp cellnet.RedisPoolOperator
var balancer balance.LoadBalancer

const ReturnGatewayCount = 3

func protectedHandleFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("uri: %s  panic: %s", r.RequestURI, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func VerifyUser(uid, tokenStr string) (newToken string, err error, errCode int) {
	//todo 验证账号
	if uid == "" || tokenStr == "" {
		err = errors.New("uid or token is empty")
		errCode = responce.UidEmpty
		return
	}

	errCode = responce.Succeed

	//生成新token
	newToken, err = token.Generate(uid)
	if err != nil {
		errCode = responce.GenTokenFailed
	}
	return
}

func getGatewayAddr() []string {
	descArray := discovery.Default.Query("gateway")

	//取负载最低的3个网关
	svcIdArray, err := balancer.LeastConnections(ReturnGatewayCount)
	if err != nil {
		if len(descArray) > 0 {
			return []string{fmt.Sprintf("%s:%d", descArray[0].Host, descArray[0].Port)}
		} else {
			return nil
		}
	}

	result := make([]string, 0, ReturnGatewayCount)

	//找出网关的ip地址
	for _, desc := range descArray {
		for index, svcID := range svcIdArray {
			if desc.ID == svcID {
				result = append(result, fmt.Sprintf("%s:%d", desc.Host, desc.Port))
				svcIdArray = append(svcIdArray[:index], svcIdArray[index+1:]...)
				break
			}
		}

		if len(svcIdArray) == 0 {
			break
		}
	}

	return result
}

func login(w http.ResponseWriter, r *http.Request) {
	uid := r.FormValue("uid")
	token := r.FormValue("token")
	//log.logger.Infof("uid=%s, token=%s", uid, token)

	newToken, err, errCode := VerifyUser(uid, token)

	var result responce.LoginResponce
	if err == nil {
		//账号验证成功
		gatewayAddrs := getGatewayAddr()
		if len(gatewayAddrs) == 0 {
			//当前没有可用的网关
			result = responce.LoginResponce{
				Ret: responce.NoAvailableGateway,
				Msg: "no available gateway",
			}
		} else {
			result = responce.LoginResponce{
				Ret: errCode,
				User: &responce.LoginData{
					Uid:         uid,
					Token:       newToken,
					GatewayAddr: gatewayAddrs,
				},
			}
		}
	} else {
		//账号验证失败
		result = responce.LoginResponce{
			Ret: errCode,
			Msg: err.Error(),
		}
	}

	data, err := json.Marshal(result)
	if err == nil {
		w.Write(data)
	} else {
		logger.Errorf("json.Marshal error: %v", err)
	}
}

func startHttp() {
	//启动http服务
	logger.Infof("start http...")
	http.Handle("/login", protectedHandleFunc(http.HandlerFunc(login)))
	//http.HandleFunc("/login", login)
	server := &http.Server{
		Addr:         ":8000",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	go server.ListenAndServe()
}

func main() {
	base.Init("login")

	//watch网关服务
	discovery.Default.WatchSvc("gateway", nil)

	balancer = balance.New("gateway")

	//启动http
	startHttp()

	base.StartLoop()
	base.Exit()
}
