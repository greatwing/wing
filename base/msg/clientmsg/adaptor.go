package clientmsg

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/msg"
	"github.com/greatwing/wing/base/service"
	"github.com/greatwing/wing/proto"
	"reflect"
)

type MsgCallback func(ev cellnet.Event, cid proto.ClientID)
type ClosedCallback func(cid proto.ClientID)

//监听客户端消息
func On(msgId proto.Ptl, callback MsgCallback) {
	msg.On(int(msgId), handleBackendMessage(callback))
}
func OnType(msgType reflect.Type, callback MsgCallback) {
	msg.OnType(msgType, handleBackendMessage(callback))
}

//监听客户端关闭
func OnClosed(callback ClosedCallback) {
	msg.On(proto.Ptl_ClientClosed, handleClosedMessage(callback))
}

//通过网关发送消息给客户端
func Send(session cellnet.Session, cid proto.ClientID, msg interface{}) {
	data, meta, err := codec.EncodeMessage(msg, nil)
	if err != nil {
		log.Errorf("EncodeMessage Error: %s", err)
		return
	}

	session.Send(&proto.Transmit{
		MsgID:    uint32(meta.ID),
		MsgData:  data,
		ClientID: cid.ID,
	})
}

func handleBackendMessage(userHandler MsgCallback) cellnet.EventCallback {

	return func(incomingEv cellnet.Event) {

		switch ev := incomingEv.(type) {
		case *RecvMsgEvent:

			var cid proto.ClientID
			cid.ID = ev.ClientID

			if gatewayCtx := service.SessionToContext(ev.Session()); gatewayCtx != nil {
				cid.SvcID = gatewayCtx.SvcID
			}

			userHandler(incomingEv, cid)
		}
	}
}

func handleClosedMessage(callback ClosedCallback) cellnet.EventCallback {
	return func(incomingEv cellnet.Event) {
		if ev, ok := incomingEv.Message().(*proto.ClientClosed); ok {
			callback(ev.ID)
		}
	}
}
