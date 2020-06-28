package discovery

type ValueMeta struct {
	Key   string
	Value []byte
}

type OperateType int

const (
	PUT    OperateType = 0
	DELETE OperateType = 1
)

type NotifyFunc func(op OperateType, data interface{})

var (
	Default Discovery
)

type Discovery interface {

	// 注册服务
	Register(*ServiceDesc) error

	// 解注册服务
	Deregister(svcid string) error

	// 根据服务名查到可用的服务
	Query(name string) (ret []*ServiceDesc)

	// 注册服务变化通知，回调在其他线程执行
	Watch(svcName string, callback NotifyFunc)

	// 重新拉取服务信息
	Refresh(name string)

	Close()
}
