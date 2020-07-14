package clientmsg

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/proto"
)

type RecvMsgEvent struct {
	Ses      cellnet.Session
	Msg      interface{}
	ClientID int64
}

func (self *RecvMsgEvent) Session() cellnet.Session {
	return self.Ses
}

func (self *RecvMsgEvent) Message() interface{} {
	return self.Msg
}

func (self *RecvMsgEvent) Reply(msg interface{}) {

	data, meta, err := codec.EncodeMessage(msg, nil)
	if err != nil {
		logger.Errorf("Reply.EncodeMessage %s", err)
		return
	}

	self.Ses.Send(&proto.Transmit{
		MsgID:    uint32(meta.ID),
		MsgData:  data,
		ClientID: self.ClientID,
	})

}
