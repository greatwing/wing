package main

import (
	"github.com/davyxu/cellnet"
	_ "github.com/davyxu/cellnet/codec/gogopb"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/golog"
	"github.com/greatwing/wing/proto"
)

var log = golog.New("client")

func main() {

	done := make(chan struct{})

	// 创建一个事件处理队列，整个客户端只有这一个队列处理事件，客户端属于单线程模型
	queue := cellnet.NewEventQueue()

	// 创建一个tcp的连接器，名称为client，连接地址为127.0.0.1:8801，将事件投递到queue队列,单线程的处理（收发封包过程是多线程）
	p := peer.NewGenericPeer("tcp.Connector", "client", "127.0.0.1:18801", queue)

	// 设定封包收发处理的模式为tcp的ltv(Length-Type-Value), Length为封包大小，Type为消息ID，Value为消息内容
	// 并使用switch处理收到的消息
	proc.BindProcessorHandler(p, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			log.Debugln("client connected")
			ev.Session().Send(&proto.UserInfo{
				Message: "I am Client",
				Length:  1,
				Cnt:     2,
			})
		case *cellnet.SessionClosed:
			log.Debugln("client error")
		case *proto.UserInfo:
			log.Infof("msg:%v, len:%v, cnt:%d", msg.Message, msg.Length, msg.Cnt)
			done <- struct{}{}
		}
	})

	// 开始发起到服务器的连接
	p.Start()

	// 事件队列开始循环
	queue.StartLoop()

	<-done
}
