package base

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/msg"
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
	ListenAddr   string // socket侦听地址,scheme://host:minPort~maxPort/path
	MaxConnCount int    // 最大连接数量
}

// 初始化框架
func Init(svcName string) {
	log.Infof("config = %s", config.ToJson())

	msglog.SetCurrMsgLogMode(msglog.MsgLogMode_BlackList)
	msglog.SetMsgLogRule("proto.PingACK", msglog.MsgLogRule_BlackList)
	msglog.SetMsgLogRule("proto.SvcStatusACK", msglog.MsgLogRule_BlackList)

	Queue = cellnet.NewEventQueue()
	Queue.StartLoop()

	service.Init(svcName)
	service.ConnectDiscovery() //连接服务发现
}

// 等待退出信号
func StartLoop() {
	service.WaitExitSignal()
}

// 退出处理
func Exit() {
	service.StopAllService()
	log.Sync()
	log.Rotate()
}

func Accept(param ServiceParameter) cellnet.Peer {

	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Acceptor"
	}

	if param.SvcName == "" {
		param.SvcName = config.GetSvcName()
	}

	p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, param.ListenAddr, Queue)

	//"svc.backend"
	proc.BindProcessorHandler(p, param.NetProcName, msg.Process)

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

// 连接指定的服务
//  filters 根据规则过滤发现的服务,FilterFunc返回true表示匹配
func Connect(param ServiceParameter, filters ...service.FilterFunc) service.MultiPeer {
	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Connector"
	}

	mp := service.DiscoveryService(param.SvcName, param.MaxConnCount, func(multiPeer service.MultiPeer, sd *discovery.ServiceDesc) {

		p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, sd.Address(), Queue)

		proc.BindProcessorHandler(p, param.NetProcName, msg.Process)

		if opt, ok := p.(cellnet.TCPSocketOption); ok {
			opt.SetSocketBuffer(2048, 2048, true)
		}

		p.(cellnet.TCPConnector).SetReconnectDuration(time.Second * 3)
		multiPeer.AddPeer(sd, p)
		p.Start()

	}, filters...)

	mp.(service.MultiPeer).SetContext("multi", param)

	service.AddLocalService(mp.(cellnet.Peer))

	return mp
}

func ConnectToRedis(addr string) {
	p := peer.NewGenericPeer("redix.Connector", "redis", addr, Queue)
	p.Start()
	service.AddLocalService(p)
}
