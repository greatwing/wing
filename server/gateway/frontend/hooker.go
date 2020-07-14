package frontend

import (
	"errors"
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service/discovery/location"
	"github.com/greatwing/wing/proto"
	"github.com/greatwing/wing/server/gateway/route"
	"github.com/greatwing/wing/server/login/token"
	"time"
)

func ProcFrontendPacket(msgID int, msgData []byte, ses cellnet.Session) (msg interface{}, err error) {
	// gateway自己的内部消息以及预处理消息
	switch msgID {
	case int(proto.Ptl_LoginReq):
		// 将字节数组和消息ID用户解出消息
		msg, _, err = codec.DecodeMessage(msgID, msgData)
		if err != nil {
			return nil, err
		}

		if req, ok := msg.(*proto.Msg_LoginReq); ok {
			//todo 验证用户token
			err = token.Verify(req.Uid, req.Token)
			if err == nil {
				logger.Infof("login succeed uid: %s", req.Uid)
			} else {
				ses.Close()
				logger.Infof("login failed uid: %s, err: %s", req.Uid, err)
				return
			}

			//选取game
			gameSvcId := location.GetUserLocation(req.Uid)
			u, err := bindClientToBackend(gameSvcId, ses.ID())
			if err == nil {
				u.TransmitToBackend(gameSvcId, msgID, msgData)
			} else {
				ses.Close()
				logger.Error("bindClientToBackend", err)
			}
		} else {
			return nil, errors.New("LoginReq decode error")
		}

	case int(proto.Ptl_Ping):
		u := SessionToClient(ses)
		if u != nil {
			u.LastPingTime = time.Now()
			//ses.Send(&proto.Msg_Ping{})
		} else {
			ses.Close()
		}
	default:
		// 在路由规则中查找消息ID是否是路由规则允许的消息
		rule := route.GetRuleByMsgID(msgID)
		if rule == nil {
			return nil, fmt.Errorf("Message not in route table, msgid: %d, call addrule ", msgID)
		}

		// 找session绑定的user
		u := SessionToClient(ses)
		if u != nil {
			// 透传到后台
			svcID := u.GetBackend(rule.SvcName)
			if err = u.TransmitToBackend(svcID, msgID, msgData); err != nil {
				logger.Warnf("TransmitToBackend %s, msg: '%d' svc: %s", err, msgID, rule.SvcName)
			}
		} else {
			//找不到user，说明还没有发送LoginReq验证登录，断开连接
			return nil, errors.New("user not login before send msg, disconnet")
		}
	}

	return
}

type FrontendEventHooker struct {
}

// 网关内部抛出的事件
func (FrontendEventHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	switch inputEvent.Message().(type) {
	case *cellnet.SessionClosed:
		// 通知后台客户端关闭
		u := SessionToClient(inputEvent.Session())
		if u != nil {
			u.BroadcastToBackends(&proto.ClientClosed{
				ID: proto.ClientID{
					ID:    inputEvent.Session().ID(),
					SvcID: config.GetLocalSvcID(),
				},
			})
		}
	}

	return inputEvent
}

// 发送到客户端的消息
func (FrontendEventHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {

	return inputEvent
}
