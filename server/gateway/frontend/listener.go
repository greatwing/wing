package frontend

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/service"
)

func Accept(param base.ServiceParameter) {

	clientListener := peer.NewGenericPeer(param.NetPeerType, param.SvcName, param.ListenAddr, nil)

	//LoginReq在io的goroutine中处理，其他协议根据路由规则转发，不需要回调
	proc.BindProcessorHandler(clientListener, param.NetProcName, nil)

	if socketOpt, ok := clientListener.(cellnet.TCPSocketOption); ok {
		// 无延迟设置缓冲
		socketOpt.SetSocketBuffer(-1, -1, true)
		//socketOpt.SetSocketDeadline(time.Second*40, time.Second*20)
	}

	clientListener.Start()
	FrontendSessionManager = clientListener.(peer.SessionManager)

	// 服务发现注册服务
	service.Register(clientListener)

	service.AddLocalService(clientListener)
}
