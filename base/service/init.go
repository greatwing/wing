package service

import (
	"github.com/greatwing/wing/base/config"
	"github.com/greatwing/wing/base/log"
	"github.com/greatwing/wing/base/service/discovery"
	"os"
	"os/signal"
	"syscall"
)

func Init(name string) {
	config.SetSvcName(name)
}

// 连接到服务发现, 建议在service.Init后, 以及服务器逻辑开始前调用
func ConnectDiscovery() {
	if config.GetDiscoveryAddr() == nil {
		logger.Panic("no discovery address")
	}
	discovery.Default = discovery.NewDiscovery(config.GetDiscoveryAddr())
}

func WaitExitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	//signal.Notify(ch)
	s := <-ch
	logger.Infof("signal %v", s)
}
