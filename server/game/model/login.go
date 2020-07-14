package model

import (
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/msg"
	"github.com/greatwing/wing/base/msg/clientmsg"
	"github.com/greatwing/wing/base/service"
	"github.com/greatwing/wing/base/service/discovery/location"
	"github.com/greatwing/wing/proto"
	"github.com/greatwing/wing/server/game/model/db"
	"github.com/greatwing/wing/server/game/model/role"
)

func onLogin(user *User, cid proto.ClientID) {

	if user.Role() == nil {
		//异步查询角色数据
		db.QueryRoleByUidAsync(user.Id, func(r *role.Role, roleErr error) {
			base.RunInLogicGoroutine(func() {
				//在逻辑线程运行
				if user.Cid != cid {
					//user的cid已发生改变，说明同一账号在其他设备上登录，已被踢下线
					return
				}
				if roleErr != nil {
					//读取出错，关闭客户端
					closeClient(cid)
					return
				}

				result := proto.Msg_LoginRsp_Succeed
				if r == nil {
					result = proto.Msg_LoginRsp_NoRole
				} else {
					user.SetRole(r)

					//todo 测试用
					db.SaveRole(user.Role())
				}
				//回复登录结果
				clientmsg.Send(cid, &proto.Msg_LoginRsp{
					Result: result,
				})
			})
		})
	} else {
		//已有角色数据，不用查询了，直接回复登录结果
		clientmsg.Send(cid, &proto.Msg_LoginRsp{
			Result: proto.Msg_LoginRsp_Succeed,
		})
	}
}

func onLogoff(cid proto.ClientID) {

	//关闭client
	closeClient(cid)

	//删除user
	user := GetUser(cid)
	if user != nil {
		logger.Infof("logoff uid=%v", user.Id)
		RemoveUser(user)

		//判断是否删除location
		msg.FireCustom("DelUserLocation", user.Id)
	}
}

func closeClient(cid proto.ClientID) {
	session := service.GetRemoteService(cid.SvcID)
	if session != nil {
		session.Send(&proto.CloseClient{
			ID: []int64{cid.ID},
		})
	}
}

func init() {
	//处理登录协议
	clientmsg.On(proto.Msg_LoginReq{}, func(ev cellnet.Event, cid proto.ClientID) {
		msg := ev.Message().(*proto.Msg_LoginReq)
		logger.Infof("user login uid:%v, token:%v", msg.Uid, msg.Token)

		//先添加user，避免同一账户同时登陆造成逻辑问题
		user := GetUserByUid(msg.Uid)
		if user != nil {
			//用户已在线，踢下线
			closeClient(user.Cid)
			RemoveUser(user)

			//修改cid
			user.Cid = cid
		} else {
			user = &User{
				Id:  msg.Uid,
				Cid: cid,
			}
		}
		AddUser(user)

		//注意不要卡住逻辑线程
		go func() {
			err := location.SetUserLocation(msg.Uid, config.GetLocalSvcID()) //阻塞
			if err != nil {
				//可能已经在其他game节点上登录了，关闭客户端
				base.RunInLogicGoroutine(func() {
					onLogoff(cid)
				})
			} else {
				base.RunInLogicGoroutine(func() {
					if user.Cid == cid {
						//cid没变，说明没有被挤下线
						onLogin(user, cid)
					}
				})
			}
		}()
	})

	//处理客户端断开连接
	clientmsg.OnClosed(func(cid proto.ClientID) {
		logger.Debugf("client closed: %v@%s", cid.ID, cid.SvcID)
		onLogoff(cid)
	})

	//处理创建角色
	clientmsg.On(proto.Msg_CreatRoleReq{}, func(ev cellnet.Event, cid proto.ClientID) {
		msg := ev.Message().(*proto.Msg_CreatRoleReq)
		logger.Infof("request to creat role, name=%s, gender=%d", msg.Name, msg.Gender)
		user := GetUser(cid)
		if user == nil {
			closeClient(cid)
			return
		}

		if user.Role() == nil {
			//创建角色
			user.SetRole(role.New(user.Id, msg.Name, msg.Gender))

			//todo 添加测试道具
			for i := 0; i < 1000; i++ {
				user.Role().NewItem(int32(i), 2)
			}

			//保存到数据库
			db.SaveRole(user.Role())

			clientmsg.Send(cid, &proto.Msg_CreatRoleRsp{
				Result: proto.Msg_CreatRoleRsp_Succeed,
			})
		} else {
			//已经有角色了(单账号单角色)
			clientmsg.Send(cid, &proto.Msg_CreatRoleRsp{
				Result: proto.Msg_CreatRoleRsp_MaxRoleCount,
			})
		}
	})

	//删除user location
	msg.OnCustom("DelUserLocation", func(param ...interface{}) {
		if len(param) < 1 {
			//参数不正确
			return
		}

		if uid, ok := param[0].(string); ok {
			if GetUserByUid(uid) == nil && !db.IsRoleCached(uid) {
				go func() {
					err := location.DelUserLocation(uid, config.GetLocalSvcID())
					if err != nil {
						logger.Error(err)
					}
				}()
			}
		}
	})
}
