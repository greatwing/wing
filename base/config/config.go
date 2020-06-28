package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	procName string
	cfg      jsonConfig
	isDebug  bool
)

type jsonConfig struct {
	SvcGroup      string         `json:"svc_group"`
	SvcIndex      int            `json:"svc_index"`
	WanIp         string         `json:"wan_ip"`
	DiscoveryAddr []string       `json:"discovery_addr"`
	Frontend      frontendConfig `json:"frontend"`
}

type frontendConfig struct {
	Protocol string `json:"protocol,omitempty"`
	MinPort  int    `json:"min_port,omitempty"`
	MaxPort  int    `json:"max_port,omitempty"`
}

// 获取当前服务进程名称
func GetSvcName() string {
	return procName
}
func SetSvcName(name string) {
	procName = name
}

// 获取外网IP
func GetWANIP() string {
	return cfg.WanIp
}

func GetSvcGroup() string {
	return cfg.SvcGroup
}

func GetSvcIndex() int {
	return cfg.SvcIndex
}

func GetDiscoveryAddr() []string {
	return cfg.DiscoveryAddr
}

func ToJson() string {
	jsonData, _ := json.Marshal(cfg)
	return string(jsonData)
}

func Debug() bool {
	return isDebug
}

func GetPortsRange() string {
	if cfg.Frontend.MinPort <= 0 {
		if cfg.Frontend.MaxPort > 0 {
			return string(cfg.Frontend.MaxPort)
		} else {
			return "0"
		}
	} else if cfg.Frontend.MinPort > cfg.Frontend.MaxPort {
		return string(cfg.Frontend.MinPort)
	} else {
		return fmt.Sprintf("%d~%d", cfg.Frontend.MinPort, cfg.Frontend.MaxPort)
	}
}

func GetProtocol() string {
	if cfg.Frontend.Protocol == "" {
		return "tcp"
	}

	return cfg.Frontend.Protocol
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

func init() {
	cfg = jsonConfig{}

	configPath := flag.String("c", "", "config file path")
	svcGroup := flag.String("g", "", "service group name")
	svcIndex := flag.Int("i", 0, "service index")
	wanIp := flag.String("w", "", "wan ip addr")
	etcdAddresses := flag.String("e", "", "etcd addresses")
	protocol := flag.String("p", "tcp", "frontend protocol")
	minPort := flag.Int("min", 0, "frontend min port")
	maxPort := flag.Int("max", 0, "frontend max port")
	flag.BoolVar(&isDebug, "d", false, "debug mode")
	flag.Parse()

	if *configPath == "" && FileExist("config.json") {
		*configPath = "config.json"
	}

	//如果设置了配置文件，则从文件中读取配置
	if len(*configPath) > 0 {
		data, err := ioutil.ReadFile(*configPath)
		if err == nil {
			jsonErr := json.Unmarshal(data, &cfg)
			if jsonErr != nil {
				panic(fmt.Sprintf("read config error: %v", jsonErr))
			}
		} else {
			fmt.Printf("cannot find config file: %s", *configPath)
		}
	}

	//命令行覆盖配置文件
	flag.Visit(func(f *flag.Flag) {
		//fmt.Printf("flag %v is set to %v\n", f.Name, f.Value)
		switch f.Name {
		case "g":
			//fmt.Printf("overwrite group from %q to %q\n", cfg.SvcGroup, *svcGroup)
			cfg.SvcGroup = *svcGroup
		case "i":
			cfg.SvcIndex = *svcIndex
		case "w":
			cfg.WanIp = *wanIp
		case "e":
			//多个etcd地址用逗号分割
			cfg.DiscoveryAddr = strings.Split(*etcdAddresses, ",")
		case "p":
			cfg.Frontend.Protocol = *protocol
		case "min":
			cfg.Frontend.MinPort = *minPort
		case "max":
			cfg.Frontend.MaxPort = *maxPort
		}
	})
}
