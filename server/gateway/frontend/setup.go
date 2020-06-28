package frontend

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/gorillaws"
	"github.com/davyxu/cellnet/proc/tcp"
)

func init() {

	// 前端的processor
	proc.RegisterProcessor("tcp.frontend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		bundle.SetTransmitter(new(directTCPTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			new(tcp.MsgHooker),       // TCP基础消息及日志
			new(FrontendEventHooker), // 内部消息处理
		))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})

	// 前端的processor
	proc.RegisterProcessor("ws.frontend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		bundle.SetTransmitter(new(directWSMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			new(gorillaws.MsgHooker), // TCP基础消息及日志
			new(FrontendEventHooker), // 内部消息处理
		))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}
