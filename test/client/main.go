package main

import (
	"encoding/json"
	"fmt"
	"github.com/davyxu/cellnet"
	_ "github.com/davyxu/cellnet/codec/gogopb"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/timer"
	"github.com/greatwing/wing/proto"
	"github.com/greatwing/wing/server/login/responce"
	"io/ioutil"
	"net/http"
	"time"
)

func login(uid, token string) (*responce.LoginResponce, error) {
	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:8000/login?uid=%s&token=%s", uid, token))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {

		return nil, err
	}

	logger.Infof("%s", data)

	rsp := &responce.LoginResponce{}
	err = json.Unmarshal(data, rsp)
	return rsp, err
}

func main() {
	logger.Infof("is debug mode? %v", config.Debug())
	logger.Infof("time now : %s", time.Now())

	rsp, err := login("test", "test")
	if err != nil {
		logger.Error(err)
		return
	}

	switch rsp.Ret {
	case responce.Succeed:
		logger.Infof("login succeed uid=%v, gateway=%s", rsp.User.Uid, rsp.User.GatewayAddr)
	default:
		logger.Infof("login failed: %s", rsp.Msg)
		return
	}

	done := make(chan struct{})

	// 创建一个事件处理队列，整个客户端只有这一个队列处理事件，客户端属于单线程模型
	queue := cellnet.NewEventQueue()

	// 创建一个tcp的连接器，名称为client，连接地址为127.0.0.1:8801，将事件投递到queue队列,单线程的处理（收发封包过程是多线程）
	p := peer.NewGenericPeer("tcp.Connector", "client", rsp.User.GatewayAddr[0], queue)

	// 设定封包收发处理的模式为tcp的ltv(Length-Type-Value), Length为封包大小，Type为消息ID，Value为消息内容
	// 并使用switch处理收到的消息
	proc.BindProcessorHandler(p, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			logger.Info("client connected")
			ev.Session().Send(&proto.Msg_LoginReq{
				Uid:   rsp.User.Uid,
				Token: rsp.User.Token,
			})
		case *cellnet.SessionClosed:
			logger.Info("client closed!! ")
		case *proto.Msg_LoginRsp:
			switch msg.Result {
			case proto.Msg_LoginRsp_Succeed:
				logger.Info("login succeed")
			case proto.Msg_LoginRsp_Failed:
				logger.Info("login failed")
			case proto.Msg_LoginRsp_NoRole:
				roleName := "test_sp"
				ev.Session().Send(&proto.Msg_CreatRoleReq{
					Name:   roleName,
					Gender: proto.Gender_Female,
				})
				logger.Infof("send creat role: %v", roleName)
			}

			//发送心跳
			t := timer.New(nil)
			t.Run(func(dt time.Duration) {
				ev.Session().Send(&proto.Msg_Ping{})
			}, time.Second*60, 0)

		case *proto.Msg_CreatRoleRsp:
			logger.Infof("creat role result: %v", msg.Result)
			//done <- struct{}{}
		}
	})

	// 开始发起到服务器的连接
	p.Start()

	// 事件队列开始循环
	queue.StartLoop()

	<-done
}
