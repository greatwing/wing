package msghandler

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/msg/clientmsg"
	"github.com/greatwing/wing/base/service/serviceid"
	"github.com/greatwing/wing/proto"
)

func init() {
	//处理登录协议
	clientmsg.On(proto.Ptl_LoginReq, func(ev cellnet.Event, cid proto.ClientID) {
		if msg, ok := ev.Message().(*proto.Msg_LoginReq); ok {
			log.Infof("user login uid:%v, token:%v", msg.Uid, msg.Token)
			clientmsg.Send(ev.Session(), cid, &proto.Msg_LoginRsp{
				Result:  proto.Msg_LoginRsp_Succeed,
				Message: fmt.Sprintf("this is %s", serviceid.GetLocalSvcID()),
			})
		} else {
			log.Errorf("decode error msgId:%d", 1)
		}
	})

	//处理客户端断开连接
	clientmsg.OnClosed(func(cid proto.ClientID) {
		log.Debugf("client closed: %v@%s", cid.ID, cid.SvcID)
	})
}
