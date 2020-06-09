package proto

import "github.com/greatwing/wing/base"

func init() {
	base.RegisterPbMsgMeta(1, (*UserInfo)(nil))
}
