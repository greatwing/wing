package base

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	_ "github.com/davyxu/cellnet/codec/gogopb"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/tcp"
	"reflect"
)

func ListenMsg(msgId int, msgType reflect.Type) {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("gogopb"),
		//Type:  reflect.TypeOf((*UserInfo)(nil)).Elem(),
		Type: msgType,
		ID:    msgId,
	})
}

func init() {

	// 避免默认消息日志显示本条消息
	msglog.SetMsgLogRule("proto.TransmitACK", msglog.MsgLogRule_BlackList)

	// 适用于后端服务的处理器
	proc.RegisterProcessor("svc.backend", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback, args ...interface{}) {

		bundle.SetTransmitter(new(tcp.TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(
			//new(service.SvcEventHooker), // 服务互联处理
			//new(BackendMsgHooker),       // 网关消息处理
			new(tcp.MsgHooker))) // tcp基础消息处理
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}
