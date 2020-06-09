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
	RegisterNotify(callback NotifyFunc)

	// 设置值
	SetValue(key string, value interface{}, optList ...interface{}) error

	// 取值，并赋值到变量
	GetValue(key string, valuePtr interface{}) error

	// 删除值
	DeleteValue(key string) error

	Close()
}
