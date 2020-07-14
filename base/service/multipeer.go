package service

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/greatwing/wing/base/service/discovery"
	"sync"
)

// 一类服务发起多个连接(不是同一地址), 比如 login1 login2
type MultiPeer interface {
	GetPeers() []cellnet.Peer

	cellnet.ContextSet

	AddPeer(sd *discovery.ServiceDesc, p cellnet.Peer)

	SetSkipCheck(skip bool)
	IsSkipCheck() bool
}

type multiPeer struct {
	peer.CoreContextSet
	peers          []cellnet.Peer
	peersGuard     sync.RWMutex
	context        interface{}
	skipReadyCheck bool
}

func (m *multiPeer) Start() cellnet.Peer {
	return m
}

func (m *multiPeer) Stop() {

}

func (m *multiPeer) TypeName() string {
	return ""
}

func (m *multiPeer) GetPeers() []cellnet.Peer {
	m.peersGuard.RLock()
	defer m.peersGuard.RUnlock()

	return m.peers
}

func (m *multiPeer) SetSkipCheck(skip bool) {
	m.skipReadyCheck = skip
}

func (m *multiPeer) IsSkipCheck() bool {
	return m.skipReadyCheck
}

func (m *multiPeer) IsReady() bool {
	if m.skipReadyCheck {
		return true
	}

	peers := m.GetPeers()

	if len(peers) == 0 {
		return false
	}

	for _, p := range peers {
		if !p.(cellnet.PeerReadyChecker).IsReady() {
			return false
		}
	}

	return true
}

// 保证AddPeer在Peer  Start之前调用, 否则在连接上时因为没有sd,会导致不汇报服务信息
func (m *multiPeer) AddPeer(sd *discovery.ServiceDesc, p cellnet.Peer) {

	contextSet := p.(cellnet.ContextSet)
	contextSet.SetContext("sd", sd)

	m.peersGuard.Lock()
	m.peers = append(m.peers, p)
	m.peersGuard.Unlock()
}

func (m *multiPeer) GetPeer(svcid string) cellnet.Peer {
	m.peersGuard.RLock()
	defer m.peersGuard.RUnlock()

	for _, p := range m.peers {

		if getSvcIDByPeer(p) == svcid {
			return p
		}
	}

	return nil
}

func (m *multiPeer) RemovePeer(svcid string) {
	m.peersGuard.Lock()
	defer m.peersGuard.Unlock()
	for index, p := range m.peers {

		if getSvcIDByPeer(p) == svcid {
			m.peers = append(m.peers[:index], m.peers[index+1:]...)
			break
		}
	}
}

func getSvcIDByPeer(p cellnet.Peer) string {
	var sd *discovery.ServiceDesc
	if p.(cellnet.ContextSet).FetchContext("sd", &sd) {
		return sd.ID
	}

	return ""
}

func newMultiPeer() *multiPeer {
	return &multiPeer{}
}
