package config

import (
	"fmt"
	"strings"
)

type RunConfig struct {
	Elastic struct {
		Host     []string `yaml:"hosts"`
		UserName string   `yaml:"username"`
		Password string   `yaml:"password"`
	} `yaml:"elastic"`
	Cache struct {
		MemoryTTLSeconds int `yaml:"memory_ttl_seconds"` // 本地内存缓存秒数，默认 5
	} `yaml:"cache"`
	RateLimit struct {
		QPS   int `yaml:"qps"`   // 单 IP 每秒令牌，默认 20
		Burst int `yaml:"burst"` // 桶容量，默认 qps*2
	} `yaml:"rate_limit"`
	// Games: gameId(string) -> symbol，可选；用于列表/详情兼容老注单
	Games          map[string]string `yaml:"games"`
	ServerPort     int               `yaml:"server_port"`
	DatacenterGrpc string            `yaml:"datacenter_grpc"`
	LotteryGrpc    string            `yaml:"lottery_grpc"`
	// InternalAPIKey 供 ClientApiRpcProd 等服务端调用内部查单接口
	InternalAPIKey string `yaml:"internal_api_key"`
	Log            string `yaml:"log"`
}

func (c *RunConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if c.ServerPort <= 0 || c.ServerPort > 65535 {
		return fmt.Errorf("server_port invalid")
	}
	if strings.TrimSpace(c.DatacenterGrpc) == "" {
		return fmt.Errorf("datacenter_grpc required")
	}
	if strings.TrimSpace(c.LotteryGrpc) == "" {
		return fmt.Errorf("lottery_grpc required")
	}
	if len(c.Elastic.Host) == 0 {
		return fmt.Errorf("elastic.hosts required")
	}
	if c.Cache.MemoryTTLSeconds < 0 {
		return fmt.Errorf("cache.memory_ttl_seconds invalid")
	}
	if c.RateLimit.QPS < 0 || c.RateLimit.Burst < 0 {
		return fmt.Errorf("rate_limit invalid")
	}
	return nil
}
