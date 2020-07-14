package frontend

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/greatwing/wing/base/msg"
	"github.com/greatwing/wing/proto"
	"github.com/greatwing/wing/server/gateway/model"
)

var (
	FrontendSessionManager peer.SessionManager
)

func GetClientSession(sesid int64) cellnet.Session {

	return FrontendSessionManager.GetSession(sesid)
}

func GetClient(sesid int64) *model.Client {
	return SessionToClient(GetClientSession(sesid))
}

// 创建一个网关用户
func CreateClient(clientSes cellnet.Session) *model.Client {

	u := model.NewClient(clientSes)

	// 绑定到session上
	clientSes.(cellnet.ContextSet).SetContext("client", u)
	return u
}

// 用session获取用户
func SessionToClient(clientSes cellnet.Session) *model.Client {

	if clientSes == nil {
		return nil
	}

	if raw, ok := clientSes.(cellnet.ContextSet).GetContext("client"); ok {
		if clt, ok := raw.(*model.Client); ok {
			return clt
		}
	}

	return nil
}

// 清除session绑定的用户
func ClearClient(clientSes cellnet.Session) {
	if clientSes == nil {
		return
	}

	if ctx, ok := clientSes.(cellnet.ContextSet); ok {
		ctx.SetContext("client", nil)
	}
}

// 遍历所有的用户
func VisitClient(callback func(*model.Client) bool) {
	FrontendSessionManager.VisitSession(func(clientSes cellnet.Session) bool {

		if u := SessionToClient(clientSes); u != nil {
			return callback(u)
		}

		return true
	})
}

func closeClient(clt *model.Client) {
	//清除session上的client，防止又把ClientClosed消息发回给后台
	ClearClient(clt.ClientSession)

	//关闭session
	clt.ClientSession.Close()
}

func init() {
	//后台要求主动关闭client
	msg.OnType(proto.CloseClient{}, func(ev cellnet.Event) {
		if closeMsg, ok := ev.Message().(*proto.CloseClient); ok {

			// 不给ID,关掉这个网关的所有客户端
			if len(closeMsg.ID) == 0 {
				VisitClient(func(clt *model.Client) bool {
					closeClient(clt)
					return true
				})

			} else {
				// 关闭指定的客户端
				for _, sesId := range closeMsg.ID {
					if clt := GetClient(sesId); clt != nil {
						closeClient(clt)
					}
				}
			}
		}
	})
}
