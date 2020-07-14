package service

import (
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service/discovery"
	"github.com/greatwing/wing/base/utils"
)

type FilterFunc func(*discovery.ServiceDesc) bool
type QueryServiceOp int
type QueryResult interface{}

// 发现一个服务，服务可能拥有多个地址，每个地址返回时，创建一个connector并开启
// DiscoveryService返回值返回持有多个Peer的peer, 判断Peer的IsReady可以得到所有连接准备好的状态
func DiscoveryService(tgtSvcName string, maxCount int,
	peerCreator func(MultiPeer, *discovery.ServiceDesc), filters ...FilterFunc) MultiPeer {

	// 从发现到连接有一个过程，需要用Map防止还没连上，又创建一个新的连接
	multiPeer := newMultiPeer()

	discovery.Default.WatchSvc(tgtSvcName, func(op discovery.OperateType, desc *discovery.ServiceDesc) {
		switch op {
		case discovery.PUT:
			if desc != nil && desc.Name == tgtSvcName {
				if Filter(desc, filters...) {

					logger.Infof("found '%s' address '%s' ", desc.Name, desc.Address())

					prePeer := multiPeer.GetPeer(desc.ID)

					// 如果svcid重复汇报, 可能svcid内容有变化
					if prePeer != nil {

						var preDesc *discovery.ServiceDesc
						if prePeer.(cellnet.ContextSet).FetchContext("sd", &preDesc) && !preDesc.Equals(desc) {

							logger.Infof("service '%s' change desc, %+v -> %+v...", desc.ID, preDesc, desc)

							// 移除之前的连接
							multiPeer.RemovePeer(desc.ID)

							// 停止重连
							prePeer.Stop()

						} else {
							//没变化，不需要创建新的peer
							return
						}

					}

					// 达到最大连接
					if maxCount > 0 && len(multiPeer.GetPeers()) >= maxCount {
						return
					}

					// 用户创建peer
					peerCreator(multiPeer, desc)
				}
			}
		case discovery.DELETE:
			if desc != nil {
				//已停止的服务不再尝试重连
				delPeer := multiPeer.GetPeer(desc.ID)
				if delPeer != nil {
					logger.Infof("close peer of svcid: %s", desc.ID)

					// 移除连接
					multiPeer.RemovePeer(desc.ID)

					// 停止重连
					delPeer.Stop()
				}
			}
		default:
			panic("[service notify] unkown operation")
		}
	})

	return multiPeer
}

// 根据filterList进行筛选
func Filter(desc *discovery.ServiceDesc, filterList ...FilterFunc) bool {
	for _, filter := range filterList {

		if filter == nil {
			continue
		}

		if !filter(desc) {
			return false
		}
	}

	return true
}

// 匹配指定的服务组,服务组空时,匹配所有
func FilterMatchSvcGroup(svcGroup string) FilterFunc {

	return func(desc *discovery.ServiceDesc) bool {

		if svcGroup == "" {
			return true
		}

		return desc.GetMeta("SvcGroup") == svcGroup
	}
}

// 匹配指定的服务ID
func FilterMatchSvcID(svcid string) FilterFunc {

	return func(desc *discovery.ServiceDesc) bool {
		return desc.ID == svcid
	}
}

// 匹配指定的规则,一般由命令行指定
func FilterMatchRule(rules []string) FilterFunc {

	return func(desc *discovery.ServiceDesc) bool {

		// 任意规则满足即可
		for _, rule := range rules {
			if utils.WildcardPatternMatch(desc.ID, rule) {
				return true
			}
		}

		return false
	}
}
