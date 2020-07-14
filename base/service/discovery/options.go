package discovery

import (
	"go.etcd.io/etcd/v3/clientv3"
)

type etcdOptionParam struct {
	key     string
	leaseID clientv3.LeaseID
	opts    []clientv3.OpOption
}

func (p *etcdOptionParam) Key() string {
	return p.key
}

func (p *etcdOptionParam) SetKey(key string) {
	p.key = key
}

func WithLease() Option {
	return func(opt OptionParam) {
		if etcdOpt, ok := opt.(*etcdOptionParam); ok {
			etcdOpt.opts = append(etcdOpt.opts, clientv3.WithLease(etcdOpt.leaseID))
		}
	}
}

func WithPrefix() Option {
	return func(opt OptionParam) {
		if etcdOpt, ok := opt.(*etcdOptionParam); ok {
			etcdOpt.opts = append(etcdOpt.opts, clientv3.WithPrefix())
		}
	}
}

// 加上"/service"前缀
func WithSvcKey() Option {
	return func(opt OptionParam) {
		opt.SetKey(GetServiceKey(opt.Key()))
	}
}

// 加上"/balance"前缀
func WithBalanceKey() Option {
	return func(opt OptionParam) {
		opt.SetKey(GetBalanceKey(opt.Key()))
	}
}

// 加上"/location"前缀
func WithLocationKey() Option {
	return func(opt OptionParam) {
		opt.SetKey(GetLocationKey(opt.Key()))
	}
}
