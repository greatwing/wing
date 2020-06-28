package etcd

import (
	"context"
	"encoding/json"
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service/discovery"
	"github.com/greatwing/wing/base/service/serviceid"
	"go.etcd.io/etcd/v3/clientv3"
	"sync"
	"time"
)

type discoveryEtcd struct {
	//etcd client
	client *clientv3.Client

	//服务发现缓存
	svcCache      map[string][]*discovery.ServiceDesc
	svcCacheGuard sync.RWMutex

	//watch回调
	notifyCallbacks map[string]discovery.NotifyFunc
	notifyQueue     *cellnet.EventQueue
	notifyGuard     sync.RWMutex
	watchCancel     context.CancelFunc
}

// 注册服务
func (d *discoveryEtcd) Register(desc *discovery.ServiceDesc) error {
	jsonBytes, err := json.Marshal(desc)
	if err != nil {
		return err
	}

	//注册需要等待完成，没有超时
	_, err = d.client.Put(context.Background(), serviceid.GetServiceKey(desc.ID), string(jsonBytes))
	if err != nil {
		log.Error(err)
	}
	return err
}

// 解注册服务
func (d *discoveryEtcd) Deregister(svcid string) error {
	//解注最多等10秒，防止进程不能关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := d.client.Delete(ctx, serviceid.GetServiceKey(svcid))
	cancel()
	if err != nil {
		log.Error(err)
	}
	return err
}

// 根据服务名查到可用的服务
func (d *discoveryEtcd) Query(name string) (ret []*discovery.ServiceDesc) {
	d.svcCacheGuard.RLock()
	defer d.svcCacheGuard.RUnlock()

	return d.svcCache[name]
}

// 注册服务变化通知
func (d *discoveryEtcd) Watch(svcName string, callback discovery.NotifyFunc) {
	d.notifyGuard.Lock()
	d.notifyCallbacks[svcName] = callback
	d.notifyGuard.Unlock()

	d.startWatch(svcName)

	// 立刻获取一下服务
	d.Refresh(svcName)
}

// 拉取所有的服务信息
func (d *discoveryEtcd) Refresh(svcName string) {
	resp, err := d.client.Get(context.Background(), serviceid.ServiceKeyPrefix+"/"+svcName, clientv3.WithPrefix())
	if err != nil {
		log.Error(err)
	} else {
		for _, ev := range resp.Kvs {
			//log.Infof("%q : %q\n", ev.Key, ev.Value)
			var desc discovery.ServiceDesc
			err := json.Unmarshal(ev.Value, &desc)
			if err == nil {
				d.updateSvcCache(&desc)
			}
		}
	}
}

//关闭etcd client
func (d *discoveryEtcd) Close() {
	if d.watchCancel != nil {
		d.watchCancel()
	}
	d.client.Close()
}

//watch service变化
func (d *discoveryEtcd) startWatch(svcName string) {
	var ctx context.Context
	ctx, d.watchCancel = context.WithCancel(context.Background())

	log.Infof("start watch %s ...", svcName)
	rch := d.client.Watch(ctx, serviceid.ServiceKeyPrefix+"/"+svcName, clientv3.WithPrefix())
	log.Infof("watch %s succeed !", svcName)

	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				//log.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)

				switch ev.Type {
				case 0:
					//put
					var desc discovery.ServiceDesc
					err := json.Unmarshal(ev.Kv.Value, &desc)
					if err == nil {
						d.updateSvcCache(&desc)
					}
				case 1:
					//delete
					svcid := serviceid.GetSvcIDByServiceKey(string(ev.Kv.Key))
					svcName, _, _, err := serviceid.ParseSvcID(svcid)
					if err == nil {
						d.deleteSvcCache(svcid, svcName)
					}
				}
			}
		}
	}()
}

func (d *discoveryEtcd) updateSvcCache(desc *discovery.ServiceDesc) {
	d.svcCacheGuard.Lock()

	list := d.svcCache[desc.Name]

	notfound := true
	for index, svc := range list {
		if svc.ID == desc.ID {
			list[index] = desc
			notfound = false
			break
		}
	}

	if notfound {
		list = append(list, desc)
	}

	d.svcCache[desc.Name] = list
	d.svcCacheGuard.Unlock()

	//通知服务发现变化
	d.notifyGuard.RLock()
	callback, ok := d.notifyCallbacks[desc.Name]
	d.notifyGuard.RUnlock()
	if callback != nil && ok {
		callback(discovery.PUT, desc)
	}
}

func (d *discoveryEtcd) deleteSvcCache(svcid, svcName string) {
	d.svcCacheGuard.Lock()
	defer d.svcCacheGuard.Unlock()

	list := d.svcCache[svcName]

	for index, svc := range list {
		if svc.ID == svcid {
			list = append(list[:index], list[index+1:]...)
			break
		}
	}

	d.svcCache[svcName] = list
}

func NewDiscovery(addrs []string) discovery.Discovery {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   addrs,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	disc := &discoveryEtcd{
		client:          cli,
		svcCache:        make(map[string][]*discovery.ServiceDesc),
		notifyCallbacks: make(map[string]discovery.NotifyFunc),
	}

	return disc
}
