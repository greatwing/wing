package backend

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/tcp"
	"github.com/greatwing/wing/base/service"
)

func init() {

	// 避免默认消息日志显示本条消息
	msglog.SetMsgLogRule("proto.Transmit", msglog.MsgLogRule_BlackList)

	// 适用于后端服务的处理器
	proc.RegisterProcessor("svc.backend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		bundle.SetTransmitter(new(TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			new(service.SvcEventHooker), // 服务互联处理
			new(BackendMsgHooker),       // 网关消息处理
			new(tcp.MsgHooker)))         // tcp基础消息处理
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})

	//适用于网关和其他后端服务
	proc.RegisterProcessor("gateway.backend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		bundle.SetTransmitter(new(TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			new(service.SvcEventHooker), // 服务互联处理
			new(broadcasterHooker),      // 网关消息处理
			new(tcp.MsgHooker)))         // tcp基础消息处理
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}
