package etcd

import (
	"context"
	"encoding/json"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/golog"
	"github.com/greatwing/wing/base/service/discovery"
	"github.com/greatwing/wing/base/service/serviceid"
	"go.etcd.io/etcd/v3/clientv3"
	"sync"
	"time"
)

var log = golog.New("etcd")

type discoveryEtcd struct {
	//etcd client
	client *clientv3.Client

	//服务发现缓存
	svcCache      map[string][]*discovery.ServiceDesc
	svcCacheGuard sync.RWMutex

	//watch回调
	notifyCallback discovery.NotifyFunc
	notifyQueue    *cellnet.EventQueue
	notifyGuard    sync.RWMutex
	watchCancel    context.CancelFunc
}

// 注册服务
func (d *discoveryEtcd) Register(desc *discovery.ServiceDesc) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	jsonBytes, err := json.Marshal(desc)
	if err != nil {
		return err
	}

	_, err = d.client.Put(ctx, serviceid.GetServiceKey(desc.ID), string(jsonBytes))
	cancel()
	if err != nil {
		log.Errorln(err)
	}
	return err
}

// 解注册服务
func (d *discoveryEtcd) Deregister(svcid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := d.client.Delete(ctx, serviceid.GetServiceKey(svcid))
	cancel()
	if err != nil {
		log.Errorln(err)
	}
	return err
}

// 根据服务名查到可用的服务
func (d *discoveryEtcd) Query(name string) (ret []*discovery.ServiceDesc) {
	return
}

// 注册服务变化通知
func (d *discoveryEtcd) RegisterNotify(callback discovery.NotifyFunc) {
	d.notifyGuard.Lock()
	d.notifyCallback = callback
	d.notifyGuard.Unlock()
}

// 设置值
func (d *discoveryEtcd) SetValue(key string, value interface{}, optList ...interface{}) error {
	return nil
}

// 取值，并赋值到变量
func (d *discoveryEtcd) GetValue(key string, valuePtr interface{}) error {
	return nil
}

// 删除值
func (d *discoveryEtcd) DeleteValue(key string) error {
	return nil
}

//关闭etcd client
func (d *discoveryEtcd) Close() {
	if d.watchCancel != nil {
		d.watchCancel()
	}
	d.client.Close()
}

//watch service变化
func (d *discoveryEtcd) startWatch() {
	var ctx context.Context
	ctx, d.watchCancel = context.WithCancel(context.Background())
	rch := d.client.Watch(ctx, "/service", clientv3.WithPrefix())

	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				log.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)

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
	callback := d.notifyCallback
	d.notifyGuard.RUnlock()
	if callback != nil {
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
		client:   cli,
		svcCache: make(map[string][]*discovery.ServiceDesc),
	}

	disc.startWatch()

	return disc
}
