package base

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	"github.com/greatwing/wing/base/service"
	"github.com/greatwing/wing/base/service/discovery"
	"time"
)

var (
	Queue cellnet.EventQueue
)

type ServiceParameter struct {
	SvcName      string // 服务名,注册到服务发现
	NetProcName  string // cellnet处理器名称
	NetPeerType  string // cellnet的PeerType
	ListenAddr   string // socket侦听地址
	MaxConnCount int    // 最大连接数量
	NoQueue      bool   // 不使用队列
}

// 初始化框架
func Init(procName string) {

	msglog.SetCurrMsgLogMode(msglog.MsgLogMode_BlackList)
	msglog.SetMsgLogRule("proto.PingACK", msglog.MsgLogRule_BlackList)
	msglog.SetMsgLogRule("proto.SvcStatusACK", msglog.MsgLogRule_BlackList)

	Queue = cellnet.NewEventQueue()
	Queue.StartLoop()

	service.Init(procName)
	service.ConnectDiscovery() //连接服务发现
}

// 等待退出信号
func StartLoop() {
	//fxmodel.CheckReady()
	//
	//if onReady != nil {
	//	cellnet.QueuedCall(fxmodel.Queue, onReady)
	//}

	service.WaitExitSignal()
}

// 退出处理
func Exit() {
	service.StopAllService()
}

func CreateCommnicateAcceptor(param ServiceParameter) cellnet.Peer {

	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Acceptor"
	}

	var q cellnet.EventQueue
	if !param.NoQueue {
		q = Queue
	}

	p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, param.ListenAddr, q)

	//"tcp.svc"
	proc.BindProcessorHandler(p, param.NetProcName, func(ev cellnet.Event) {

		meta := cellnet.MessageMetaByMsg(ev.Message())
		if meta != nil {
			if listeners, ok := listenerByID[meta.ID]; ok {
				for _, callback := range listeners {
					callback(ev)
				}
			}
		}
	})

	if opt, ok := p.(cellnet.TCPSocketOption); ok {
		opt.SetSocketBuffer(2048, 2048, true)
	}

	//放入本地记录
	service.AddLocalService(p)

	p.Start()

	//向etcd注册
	service.Register(p)

	return p
}

func CreateCommnicateConnector(param ServiceParameter) {
	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Connector"
	}

	//msgFunc := proto.GetMessageHandler(service.GetProcName())

	opt := service.DiscoveryOption{
		MaxCount: param.MaxConnCount,
	}

	//opt.Rules = service.LinkRules

	var q cellnet.EventQueue
	if !param.NoQueue {
		q = Queue
	}

	mp := service.DiscoveryService(param.SvcName, opt, func(multiPeer service.MultiPeer, sd *discovery.ServiceDesc) {

		p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, sd.Address(), q)

		proc.BindProcessorHandler(p, param.NetProcName, func(ev cellnet.Event) {

			//if msgFunc != nil {
			//	msgFunc(ev)
			//}
		})

		if opt, ok := p.(cellnet.TCPSocketOption); ok {
			opt.SetSocketBuffer(2048, 2048, true)
		}

		p.(cellnet.TCPConnector).SetReconnectDuration(time.Second * 3)

		//
		multiPeer.AddPeer(sd, p)

		p.Start()
	})

	mp.(service.MultiPeer).SetContext("multi", param)

	service.AddLocalService(mp)

}
