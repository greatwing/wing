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
	"github.com/greatwing/wing/base/service/discovery"
	"github.com/greatwing/wing/server/login/responce"
	"net/http"
	"time"
)

func protectedHandleFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("uri: %s  panic: %s", r.RequestURI, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func VerifyUser(uid, token string) (newToken string, err error, errCode int) {
	//todo 验证账号
	if uid == "" || token == "" {
		err = errors.New("uid or token is empty")
		errCode = responce.UidEmpty
		return
	}

	errCode = responce.Succeed
	newToken = token //todo 生成新token，写入redis
	return
}

func getGatewayAddr() string {
	ret := discovery.Default.Query("gateway")

	//todo 负载均衡
	if len(ret) > 0 {
		return fmt.Sprintf("%s:%d", ret[0].Host, ret[0].Port)
	}
	return ""
}

func login(w http.ResponseWriter, r *http.Request) {
	uid := r.FormValue("uid")
	token := r.FormValue("token")
	//log.Infof("uid=%s, token=%s", uid, token)

	newToken, err, errCode := VerifyUser(uid, token)

	var result responce.LoginResponce
	if err == nil {
		//账号验证成功
		gatewayAddr := getGatewayAddr() //todo
		if gatewayAddr == "" {
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
					GatewayAddr: gatewayAddr,
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
		log.Errorf("json.Marshal error: %v", err)
	}
}

func startHttp() {
	//启动http服务
	log.Infof("start http...")
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
	discovery.Default.Watch("gateway", nil)

	//连接redis
	base.ConnectToRedis("127.0.0.1:6379")

	//当redis连接成功的时候启动http
	base.CheckReady(func() {
		startHttp()
	})

	base.StartLoop()
	base.Exit()
}
