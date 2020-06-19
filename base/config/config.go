package config

var (
	procName string
)

// 获取当前服务进程名称
func GetProcName() string {
	return procName
}
func SetProcName(name string) {
	procName = name
}

// 获取外网IP
func GetWANIP() string {
	//todo
	return "192.168.0.195"
}

func GetSvcGroup() string {
	//todo
	return "dev"
}

func GetSvcIndex() int {
	//todo
	return 0
}

func GetDiscoveryAddr() []string {
	//todo
	return []string{"127.0.0.1:2379"}
}
