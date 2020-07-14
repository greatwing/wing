package model

import (
	"fmt"
	"github.com/greatwing/wing/proto"
	"github.com/greatwing/wing/server/game/model/role"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
)

var (
	userMap   = make(map[string]*User)              // uid, User pointer
	clientMap = make(map[string]*User)              // ClientIdToStr(cid), User pointer
	roleIdMap = make(map[primitive.ObjectID]string) //roleId, uid
	userGuard sync.RWMutex
)

type User struct {
	Id  string
	Cid proto.ClientID

	role *role.Role //单账号单角色
	//Roles []*role.role //单账号多角色，角色列表
}

func (u *User) Role() *role.Role {
	return u.role
}

func (u *User) SetRole(r *role.Role) {
	userGuard.RLock()
	defer userGuard.RUnlock()

	u.role = r
	roleIdMap[r.ID] = u.Id
}

func GetUserByUid(uid string) *User {
	userGuard.RLock()
	defer userGuard.RUnlock()

	user, ok := userMap[uid]
	if !ok {
		return nil
	}
	return user
}

func GetUser(cid proto.ClientID) *User {
	userGuard.RLock()
	defer userGuard.RUnlock()

	user, ok := clientMap[ClientIdToStr(cid)]
	if !ok {
		return nil
	}
	return user
}

func GetRole(cid proto.ClientID) *role.Role {
	user := GetUser(cid)
	if user != nil {
		return user.role
	} else {
		return nil
	}
}

func GetRoleByRid(roleId primitive.ObjectID) *role.Role {
	userGuard.RLock()
	defer userGuard.RUnlock()

	uid, ok := roleIdMap[roleId]
	if !ok {
		//找不到uid
		return nil
	}

	user, ok := userMap[uid]
	if !ok {
		//找不到user
		return nil
	}

	return user.role
}

func AddUser(user *User) {
	userGuard.Lock()
	defer userGuard.Unlock()

	userMap[user.Id] = user
	clientMap[ClientIdToStr(user.Cid)] = user
}

func RemoveUser(user *User) {
	userGuard.Lock()
	defer userGuard.Unlock()

	delete(userMap, user.Id)
	delete(clientMap, ClientIdToStr(user.Cid))
}

func ClientIdToStr(cid proto.ClientID) string {
	return fmt.Sprintf("%d@%s", cid.ID, cid.SvcID)
}
