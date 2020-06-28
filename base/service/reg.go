package service

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/util"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service/discovery"
	"github.com/greatwing/wing/base/service/serviceid"
	"strconv"
)

type peerListener interface {
	Port() int
}

type ServiceMeta map[string]string

// 将Acceptor注册到服务发现,IP自动取本地IP
func Register(p cellnet.Peer, options ...interface{}) *discovery.ServiceDesc {
	host := util.GetLocalIP()

	property := p.(cellnet.PeerProperty)

	sd := &discovery.ServiceDesc{
		ID:   serviceid.MakeLocalSvcID(property.Name()),
		Name: property.Name(),
		Host: host,
		Port: p.(peerListener).Port(),
	}

	sd.SetMeta("SvcGroup", config.GetSvcGroup())
	sd.SetMeta("SvcIndex", strconv.Itoa(config.GetSvcIndex()))

	for _, opt := range options {

		switch optValue := opt.(type) {
		case ServiceMeta:
			for metaKey, metaValue := range optValue {
				sd.SetMeta(metaKey, metaValue)
			}
		}
	}

	if config.GetWANIP() != "" {
		sd.SetMeta("WANAddress", util.JoinAddress(config.GetWANIP(), sd.Port))
	}

	log.Debugf("service '%s' listen at port: %d", sd.ID, sd.Port)

	p.(cellnet.ContextSet).SetContext("sd", sd)

	// 有同名的要先解除注册，再注册，防止watch不触发
	//discovery.Default.Deregister(sd.ID)

	//注册失败时需要重试
	for {
		err := discovery.Default.Register(sd)
		if err != nil {
			log.Errorf("service register failed, %s %s", sd.String(), err.Error())
		} else {
			break
		}
	}

	return sd
}

// 解除peer注册
func Unregister(p cellnet.Peer) {
	if property, ok := p.(cellnet.PeerProperty); ok {
		id := serviceid.MakeLocalSvcID(property.Name())
		log.Debugf("Unregister %s", id)
		discovery.Default.Deregister(id)
	}
}
