package proto

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	_ "github.com/davyxu/cellnet/codec/gogopb"
	"reflect"
)

func RegisterPbMsgMeta(msgId Ptl, msg interface{}) {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("gogopb"),
		Type:  reflect.TypeOf(msg),
		ID:    int(msgId),
	})
}

func init() {
	RegisterPbMsgMeta(Ptl_Ping, (*Msg_Ping)(nil))
	RegisterPbMsgMeta(Ptl_LoginReq, (*Msg_LoginReq)(nil))
	RegisterPbMsgMeta(Ptl_LoginRsp, (*Msg_LoginRsp)(nil))
	RegisterPbMsgMeta(Ptl_CreatRoleReq, (*Msg_CreatRoleReq)(nil))
	RegisterPbMsgMeta(Ptl_CreatRoleRsp, (*Msg_CreatRoleRsp)(nil))
}
