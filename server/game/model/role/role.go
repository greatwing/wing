package role

import (
	"github.com/greatwing/wing/proto"
	"github.com/greatwing/wing/server/game/model/component"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Role struct {
	ID  primitive.ObjectID `bson:"_id,omitempty"`
	Uid string             `bson:"uid,omitempty"`

	//角色名
	Name   string       `bson:"name,omitempty"`
	Gender proto.Gender `bson:"gender,omitempty"`

	//创建时间
	CreatTime int64 `bson:"create_time,omitempty"`

	//更新时间
	UpdateTime int64 `bson:"update_time,omitempty"`

	//背包
	component.BagComponent `bson:"bag,omitempty,inline"`
}

func (r *Role) PbData() *proto.RoleData {
	return &proto.RoleData{
		Id:     r.ID.Hex(),
		Name:   r.Name,
		Gender: r.Gender,
		Items:  r.Items,
	}
}

func New(uid, name string, gender proto.Gender) *Role {
	now := time.Now().Unix()

	r := &Role{
		ID:         primitive.NewObjectID(),
		Uid:        uid,
		Name:       name,
		Gender:     gender,
		CreatTime:  now,
		UpdateTime: now,
	}

	r.BagComponent.RebuildItemMap()

	return r
}
