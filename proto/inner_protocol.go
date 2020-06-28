package proto

import "fmt"

const (
	_ = iota
	Ptl_Serviceidentify
	Ptl_CloseClient
	Ptl_ClientClosed
	Ptl_Transmit
)

// 通知其他服务器自己是什么
type ServiceIdentify struct {
	SvcName string
	SvcID   string
}

func (self *ServiceIdentify) String() string { return fmt.Sprintf("%+v", *self) }

type ClientID struct {
	ID    int64  // 客户端在网关上的SessionID
	SvcID string // 客户端在哪个网关
}

// 服务器切断客户端连接
type CloseClient struct {
	ID  []int64
	All bool
}

// 客户端主公断开连接
type ClientClosed struct {
	ID ClientID
}

// 在网关和后端服务器之间传输协议
type Transmit struct {
	MsgID        uint32  // 用户消息ID
	MsgData      []byte  // 用户消息数据
	ClientID     int64   // 单发
	ClientIDList []int64 // 列表发
	All          bool    // 全发
}
