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
	"github.com/greatwing/wing/proto"
	"github.com/greatwing/wing/server/login/responce"
	"io/ioutil"
	"net/http"
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

	log.Infof("%s", data)

	rsp := &responce.LoginResponce{}
	err = json.Unmarshal(data, rsp)
	return rsp, err
}

func main() {
	log.Infof("is debug mode? %v", config.Debug())

	rsp, err := login("test", "test")
	if err != nil {
		log.Error(err)
		return
	}

	switch rsp.Ret {
	case responce.Succeed:
		log.Infof("login succeed uid=%v, gateway=%s", rsp.User.Uid, rsp.User.GatewayAddr)
	default:
		log.Infof("login failed: %s", rsp.Msg)
		return
	}

	done := make(chan struct{})

	// 创建一个事件处理队列，整个客户端只有这一个队列处理事件，客户端属于单线程模型
	queue := cellnet.NewEventQueue()

	// 创建一个tcp的连接器，名称为client，连接地址为127.0.0.1:8801，将事件投递到queue队列,单线程的处理（收发封包过程是多线程）
	p := peer.NewGenericPeer("tcp.Connector", "client", rsp.User.GatewayAddr, queue)

	// 设定封包收发处理的模式为tcp的ltv(Length-Type-Value), Length为封包大小，Type为消息ID，Value为消息内容
	// 并使用switch处理收到的消息
	proc.BindProcessorHandler(p, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			log.Debug("client connected")
			ev.Session().Send(&proto.Msg_LoginReq{
				Uid:   rsp.User.Uid,
				Token: rsp.User.Token,
			})
		case *cellnet.SessionClosed:
			log.Debug("client error")
		case *proto.Msg_LoginRsp:
			log.Infof("login result=%d, msg=%s", msg.Result, msg.Message)
			done <- struct{}{}
		}
	})

	// 开始发起到服务器的连接
	p.Start()

	// 事件队列开始循环
	queue.StartLoop()

	<-done
}
