package service

import (
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service/discovery"
	"github.com/greatwing/wing/proto"
)

// 服务互联消息处理
type SvcEventHooker struct {
}

func (SvcEventHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	switch msg := inputEvent.Message().(type) {
	case *proto.ServiceIdentify:

		if pre := GetRemoteService(msg.SvcID); pre == nil {

			// 添加连接上来的对方服务
			AddRemoteService(inputEvent.Session(), msg.SvcID, msg.SvcName)
		}
	case *cellnet.SessionConnected:

		ctx := inputEvent.Session().Peer().(cellnet.ContextSet)

		var sd *discovery.ServiceDesc
		if ctx.FetchContext("sd", &sd) {

			// 用Connector的名称（一般是ProcName）让远程知道自己是什么服务，用于网关等需要反向发送消息的标识
			inputEvent.Session().Send(&proto.ServiceIdentify{
				SvcName: config.GetSvcName(),
				SvcID:   config.GetLocalSvcID(),
			})

			AddRemoteService(inputEvent.Session(), sd.ID, sd.Name)
		} else {

			logger.Errorf("Make sure call multi.AddPeer before peer.Start, peer: %s", inputEvent.Session().Peer().TypeName())
		}

	case *cellnet.SessionClosed:

		RemoveRemoteService(inputEvent.Session())
	}

	return inputEvent

}

func (SvcEventHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	return inputEvent
}
