package model

import (
	"errors"
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/service"
	"github.com/greatwing/wing/proto"
	"time"
)

type Backend struct {
	SvcName string
	SvcID   string // 只保留绑定后台的svcid,即便后台更换session,也无需同步
}

type Client struct {
	ClientSession cellnet.Session
	Targets       []*Backend
	LastPingTime  time.Time

	CID proto.ClientID
}

// 广播到这个用户绑定的所有后台
func (self *Client) BroadcastToBackends(msg interface{}) {

	for _, t := range self.Targets {

		backendSes := service.GetRemoteService(t.SvcID)
		if backendSes != nil {
			backendSes.Send(msg)
		}
	}
}

var (
	ErrBackendNotFound = errors.New("backend not found")
)

func (self *Client) TransmitToBackend(backendSvcid string, msgID int, msgData []byte) error {

	backendSes := service.GetRemoteService(backendSvcid)

	if backendSes == nil {
		return ErrBackendNotFound
	}

	backendSes.Send(&proto.Transmit{
		MsgID:    uint32(msgID),
		MsgData:  msgData,
		ClientID: self.CID.ID,
	})

	return nil
}

// 绑定用户后台
func (self *Client) SetBackend(svcName string, svcID string) {

	for _, t := range self.Targets {
		if t.SvcName == svcName {
			t.SvcID = svcID
			return
		}
	}

	self.CID = proto.ClientID{
		ID:    self.ClientSession.ID(),
		SvcID: config.GetLocalSvcID(),
	}

	self.Targets = append(self.Targets, &Backend{
		SvcName: svcName,
		SvcID:   svcID,
	})
}

func (self *Client) GetBackend(svcName string) string {

	for _, t := range self.Targets {
		if t.SvcName == svcName {
			return t.SvcID
		}
	}

	return ""
}

func NewClient(clientSes cellnet.Session) *Client {
	return &Client{
		ClientSession: clientSes,
	}
}
