package discovery

import (
	"context"
	"encoding/json"
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"go.etcd.io/etcd/v3/clientv3"
	"sync"
	"time"
)

const keepAliveTTL = 120

type discoveryEtcd struct {
	//etcd client
	client  *clientv3.Client
	leaseID clientv3.LeaseID

	//服务发现缓存
	svcCache      map[string][]*ServiceDesc
	svcCacheGuard sync.RWMutex

	//watch回调
	notifyCallbacks map[string]WatchSvcCallback
	notifyQueue     *cellnet.EventQueue
	notifyGuard     sync.RWMutex
}

// 注册服务
func (d *discoveryEtcd) Register(desc *ServiceDesc) error {
	jsonBytes, err := json.Marshal(desc)
	if err != nil {
		return err
	}

	//注册需要等待完成，没有超时
	return d.SetValue(desc.ID, string(jsonBytes), WithSvcKey(), WithLease())
}

// 解注册服务
func (d *discoveryEtcd) Deregister(svcid string) error {
	//解注最多等10秒，防止进程不能关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	//解除租约
	_, err := d.client.Revoke(ctx, d.leaseID)

	cancel()
	if err != nil {
		logger.Error(err)
	}
	return err
}

// 根据服务名查到可用的服务
func (d *discoveryEtcd) Query(name string) (ret []*ServiceDesc) {
	d.svcCacheGuard.RLock()
	defer d.svcCacheGuard.RUnlock()

	return d.svcCache[name]
}

// 注册服务变化通知
func (d *discoveryEtcd) WatchSvc(svcName string, callback WatchSvcCallback) {
	d.notifyGuard.Lock()
	d.notifyCallbacks[svcName] = callback
	d.notifyGuard.Unlock()

	//处理参数
	param := newParam(svcName, d.leaseID, WithSvcKey(), WithPrefix())

	d.startWatch(param, d.procSvcNotify())

	// 立刻获取一下服务
	d.refresh(param)
}

func (d *discoveryEtcd) WatchKey(key string, callback WatchCallback, opts ...Option) {
	param := newParam(key, d.leaseID, opts...)
	d.startWatch(param, callback)

	//立刻获取一下值
	resp, err := d.client.Get(context.Background(), param.key, param.opts...)
	if err != nil {
		logger.Error(err)
	} else {
		for _, ev := range resp.Kvs {
			callback(PUT, string(ev.Key), string(ev.Value))
		}
	}
}

// 设置值
func (d *discoveryEtcd) SetValue(key string, value string, opts ...Option) error {
	param := newParam(key, d.leaseID, opts...)
	_, err := d.client.Put(context.Background(), param.key, value, param.opts...)
	if err != nil {
		logger.Errorf("etcd SetValue error: %v", err)
	}
	return err
}

func (d *discoveryEtcd) GetValue(key string, opts ...Option) (string, error) {
	param := newParam(key, d.leaseID, opts...)
	rsp, err := d.client.Get(context.Background(), param.key, param.opts...)
	if err != nil {
		logger.Errorf("etcd GetValue error: %v", err)
		return "", err
	}

	var result string
	if len(rsp.Kvs) > 0 {
		result = string(rsp.Kvs[0].Value)
	}
	return result, nil
}

func (d *discoveryEtcd) DelValue(key string, opts ...Option) error {
	param := newParam(key, d.leaseID, opts...)
	_, err := d.client.Delete(context.Background(), param.key, param.opts...)
	if err != nil {
		logger.Errorf("etcd DelValue error: %v", err)
	}
	return err
}

func (d *discoveryEtcd) CheckAndSet(key string, value string, opts ...Option) (result string, err error) {
	param := newParam(key, d.leaseID, opts...)
	txn := d.client.Txn(context.Background())
	var rsp *clientv3.TxnResponse
	rsp, err = txn.If(clientv3.Compare(clientv3.CreateRevision(param.key), "=", 0)).
		Then(clientv3.OpPut(param.key, value, param.opts...)).
		Else(clientv3.OpGet(param.key, param.opts...)).
		Commit()
	if err != nil {
		return
	} else {
		if rsp.Succeeded {
			//原来没有值，设置成功
			result = value
		} else {
			//原先有值，返回原来的值
			if len(rsp.Responses) > 0 &&
				rsp.Responses[0].GetResponseRange() != nil &&
				len(rsp.Responses[0].GetResponseRange().Kvs) > 0 {

				result = string(rsp.Responses[0].GetResponseRange().Kvs[0].Value)
			}
		}
		return
	}
}

func (d *discoveryEtcd) IfDel(ctx context.Context, key string, condition string, opts ...Option) (err error) {
	param := newParam(key, d.leaseID, opts...)
	txn := d.client.Txn(ctx)
	_, err = txn.If(clientv3.Compare(clientv3.Value(param.key), "=", condition)).
		Then(clientv3.OpDelete(param.key, param.opts...)).
		Commit()
	return
}

//关闭etcd client
func (d *discoveryEtcd) Close() {
	d.client.Close()
}

// 拉取所有的服务信息
func (d *discoveryEtcd) refresh(param *etcdOptionParam) {
	resp, err := d.client.Get(context.Background(), param.key, param.opts...)
	if err != nil {
		logger.Error(err)
	} else {
		for _, ev := range resp.Kvs {
			//log.logger.Infof("%q : %q\n", ev.Key, ev.Value)
			var desc ServiceDesc
			err := json.Unmarshal(ev.Value, &desc)
			if err == nil {
				d.updateSvcCache(&desc)
			}
		}
	}
}

func (d *discoveryEtcd) procSvcNotify() WatchCallback {
	return func(op OperateType, key string, value string) {
		switch op {
		case PUT:
			var desc ServiceDesc
			err := json.Unmarshal([]byte(value), &desc)
			if err == nil {
				d.updateSvcCache(&desc)
			}
		case DELETE:
			svcid := GetSvcIDByServiceKey(key)
			svcName, _, _, err := config.ParseSvcID(svcid)
			if err == nil {
				d.deleteSvcCache(svcid, svcName)
			}
		}
	}
}

func (d *discoveryEtcd) startWatch(param *etcdOptionParam, callback WatchCallback) {
	if callback == nil {
		return
	}

	logger.Infof("start watch %s ...", param.key)
	rch := d.client.Watch(context.Background(), param.key, param.opts...)
	logger.Infof("watch %s succeed !", param.key)

	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				//log.logger.Debugf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)

				switch ev.Type {
				case 0:
					//put
					callback(PUT, string(ev.Kv.Key), string(ev.Kv.Value))
				case 1:
					//delete
					callback(DELETE, string(ev.Kv.Key), "")
				}
			}
		}
	}()
}

func (d *discoveryEtcd) updateSvcCache(desc *ServiceDesc) {
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
		callback(PUT, desc)
	}
}

func (d *discoveryEtcd) deleteSvcCache(svcid, svcName string) {
	d.svcCacheGuard.Lock()
	defer d.svcCacheGuard.Unlock()

	list := d.svcCache[svcName]

	var deletedDesc *ServiceDesc = nil
	for index, svc := range list {
		if svc.ID == svcid {
			deletedDesc = svc
			list = append(list[:index], list[index+1:]...)
			break
		}
	}

	d.svcCache[svcName] = list

	//通知服务发现变化
	d.notifyGuard.RLock()
	callback, ok := d.notifyCallbacks[svcName]
	d.notifyGuard.RUnlock()
	if callback != nil && ok {
		callback(DELETE, deletedDesc)
	}
}

// 创建参数
func newParam(key string, leaseId clientv3.LeaseID, opts ...Option) *etcdOptionParam {
	param := &etcdOptionParam{
		key:     key,
		leaseID: leaseId,
	}
	for _, opt := range opts {
		opt(param)
	}
	return param
}

func NewDiscovery(addrs []string) Discovery {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   addrs,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		logger.Fatalf("etcd new client error: %v", err)
	}

	//创建租约
	resp, err := cli.Grant(context.Background(), keepAliveTTL)
	if err != nil {
		logger.Fatalf("etcd grant error: %v", err)
	}

	ch, err := cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		logger.Fatalf("etcd keepalive error: %v", err)
	}

	go func() {
		for {
			for rsp := range ch {
				logger.Debugf("etcd keep alive ttl: %v", rsp.TTL)
			}

			//如果和etcd断开连接了，需要再次调用KeepAlive
			var kaerr error
			ch, kaerr = cli.KeepAlive(context.Background(), resp.ID)
			if kaerr != nil {
				if _, ok := kaerr.(clientv3.ErrKeepAliveHalted); !ok {
					logger.Errorf("etcd keepalive error: %v", kaerr)
				}
				break
			}
		}
	}()

	disc := &discoveryEtcd{
		client:          cli,
		leaseID:         resp.ID,
		svcCache:        make(map[string][]*ServiceDesc),
		notifyCallbacks: make(map[string]WatchSvcCallback),
	}

	return disc
}
