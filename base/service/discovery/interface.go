package discovery

import "context"

type ValueMeta struct {
	Key   string
	Value []byte
}

type OptionParam interface {
	Key() string
	SetKey(key string)
}

type OperateType int
type Option func(param OptionParam)

const (
	PUT    OperateType = 0
	DELETE OperateType = 1
)

type WatchSvcCallback func(op OperateType, desc *ServiceDesc)
type WatchCallback func(op OperateType, key string, value string)

var (
	Default Discovery
)

type Discovery interface {

	// 注册服务
	Register(*ServiceDesc) error

	// 解注册服务
	Deregister(svcid string) error

	// 根据服务名查到可用的服务
	Query(svcName string) (ret []*ServiceDesc)

	// 监控服务变化，回调在其他线程执行
	WatchSvc(svcName string, callback WatchSvcCallback)

	// 设置值
	SetValue(key string, value string, opts ...Option) error

	// 获取值
	GetValue(key string, opts ...Option) (string, error)

	// 删除值
	DelValue(key string, opts ...Option) error

	// 如果key没有值，则设置value，否则返回原来的值
	CheckAndSet(key string, value string, opts ...Option) (result string, err error)

	// 如果key的值等于condition，则删除key
	IfDel(ctx context.Context, key string, condition string, opts ...Option) (err error)

	// 监控某个key
	WatchKey(key string, callback WatchCallback, opts ...Option)

	Close()
}
