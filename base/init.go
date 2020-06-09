package base

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	"os"
	"os/signal"
	"syscall"
)

var (
	queue cellnet.EventQueue
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

	queue = cellnet.NewEventQueue()
	queue.StartLoop()
}

// 等待退出信号
func StartLoop() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
}

func CreateCommnicateAcceptor(param ServiceParameter) cellnet.Peer {

	if param.NetPeerType == "" {
		param.NetPeerType = "tcp.Acceptor"
	}

	var q cellnet.EventQueue
	if !param.NoQueue {
		q = queue
	}

	p := peer.NewGenericPeer(param.NetPeerType, param.SvcName, param.ListenAddr, q)

	//"tcp.svc"
	proc.BindProcessorHandler(p, param.NetProcName, func(ev cellnet.Event) {
		//todo
	})

	if opt, ok := p.(cellnet.TCPSocketOption); ok {
		opt.SetSocketBuffer(2048, 2048, true)
	}

	p.Start()

	return p
}
