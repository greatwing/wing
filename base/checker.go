package base

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service"
	"github.com/greatwing/wing/base/service/discovery"
	"strings"
	"time"
)

func getPeerStatus(svc cellnet.Peer) string {

	type myPeer interface {
		cellnet.PeerReadyChecker
		Name() string
		Address() string
		cellnet.Peer
	}
	mp := svc.(myPeer)

	var ready string
	if mp.IsReady() {
		ready = "READY"
	}

	var peerName string
	var context string
	if cs, ok := svc.(cellnet.ContextSet); ok {

		var desc *discovery.ServiceDesc
		if cs.FetchContext("sd", &desc) {
			context = fmt.Sprintf("--> %22s %22s", desc.ID, desc.Address())
			peerName = desc.Name
		} else {
			context = mp.Address()
			peerName = mp.Name()
		}
	}

	return fmt.Sprintf("%13s %15s %s  [%s]", peerName, mp.TypeName(), context, ready)
}

func MultiPeerString(ms service.MultiPeer) string {

	raw, ok := ms.GetContext("multi")
	if !ok {
		return ""
	}

	param := raw.(ServiceParameter)

	return fmt.Sprintf("%13s %15s", param.SvcName, param.NetPeerType)
}

func LocalServiceStatus() string {

	var sb strings.Builder

	service.VisitLocalService(func(svc cellnet.Peer) bool {

		if pg, ok := svc.(service.MultiPeer); ok {

			if !pg.IsSkipCheck() {
				// 没有连接发现时
				if len(pg.GetPeers()) == 0 {
					sb.WriteString(MultiPeerString(pg))
					sb.WriteString("\n")
				} else {
					for _, p := range pg.GetPeers() {
						sb.WriteString(getPeerStatus(p))
						sb.WriteString("\n")
					}
				}
			}

		} else {
			sb.WriteString(getPeerStatus(svc))
			sb.WriteString("\n")
		}

		return true
	})

	return sb.String()
}

func IsAllReady() (ret bool) {
	ret = true
	service.VisitLocalService(func(svc cellnet.Peer) bool {
		if !svc.(cellnet.PeerReadyChecker).IsReady() {
			ret = false
			return false
		}

		return true
	})

	return
}

//等待所有的peer都处于ready状态，阻塞
func WaitReady() {
	time.Sleep(time.Millisecond * 200)

	var lastStatus string
	for {

		if IsAllReady() {
			logger.Infof("All peers ready!\n%s", LocalServiceStatus())

			break
		}

		thisStatus := LocalServiceStatus()

		if lastStatus != thisStatus {
			logger.Infof("peers not all ready\n%s", thisStatus)
			lastStatus = thisStatus
		}

		time.Sleep(time.Second * 2) //过2秒再检测
	}
}

//检查所有的peer是否都处于ready状态，非阻塞
func CheckReady(onReady func()) {
	go func() {
		WaitReady()
		if onReady != nil {
			RunInLogicGoroutine(onReady)
		}
	}()
}
