package backend

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/tcp"
)

func init() {

	// 避免默认消息日志显示本条消息
	//msglog.SetMsgLogRule("proto.TransmitACK", msglog.MsgLogRule_BlackList)

	proc.RegisterProcessor("agent.backend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		bundle.SetTransmitter(new(tcp.TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			//new(service.SvcEventHooker), // 服务互联处理
			//new(broadcasterHooker),      // 网关消息处理
			new(tcp.MsgHooker)))         // tcp基础消息处理
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}

