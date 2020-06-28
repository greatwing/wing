package proto

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	_ "github.com/greatwing/wing/base/codec/msgp"
	"reflect"
)

func RegisterMsgpMsgMeta(msgId Ptl, msg interface{}) {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("msgp"),
		Type:  reflect.TypeOf(msg),
		ID:    int(msgId),
	})
}

func init() {
	RegisterMsgpMsgMeta(Ptl_Serviceidentify, (*ServiceIdentify)(nil))
	RegisterMsgpMsgMeta(Ptl_CloseClient, (*CloseClient)(nil))
	RegisterMsgpMsgMeta(Ptl_ClientClosed, (*ClientClosed)(nil))
	RegisterMsgpMsgMeta(Ptl_Transmit, (*Transmit)(nil))
}
