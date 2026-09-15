package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"client-api/cache"
	"client-api/common"
	"client-api/config"
	"client-api/controller"
	"client-api/dao"
	"client-api/middleware"
	"client-api/rpc"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	rootCmd = &cobra.Command{
		Use:  "client-api",
		Long: "client-api: 面向客户端的鉴权与注单查询服务",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "启动 client-api",
		Run:   run,
	}
	RunConfigPath string
)

func init() {
	runCmd.Flags().StringVar(&RunConfigPath, "config", "./config.yaml", "指定配置文件")
	rootCmd.AddCommand(runCmd)
}

func loadConfig() *config.RunConfig {
	raw, err := os.ReadFile(RunConfigPath)
	if err != nil {
		zap.L().Fatal("读取配置失败", zap.Error(err))
	}
	cfg := &config.RunConfig{}
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		zap.L().Fatal("解析配置失败", zap.Error(err))
	}
	if err := cfg.Validate(); err != nil {
		zap.L().Fatal("配置校验失败", zap.Error(err))
	}
	return cfg
}

func run(_ *cobra.Command, _ []string) {
	cfg := loadConfig()
	if err := dao.InitES(cfg); err != nil {
		zap.L().Fatal("初始化 ES 失败", zap.Error(err))
	}
	if err := dao.ES().Ping(); err != nil {
		zap.L().Fatal("ES 自检失败", zap.Error(err))
	}

	dc := rpc.NewDataCenterClient(cfg.DatacenterGrpc)
	defer dc.Close()
	lottery := rpc.NewLotteryClient(cfg.LotteryGrpc)
	defer lottery.Close()
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := dc.Ping(ctx); err != nil {
			cancel()
			zap.L().Fatal("data-center 自检失败", zap.Error(err))
		}
		if err := lottery.Ping(ctx); err != nil {
			cancel()
			zap.L().Fatal("lottery 自检失败", zap.Error(err))
		}
		cancel()
	}

	guard := cache.NewGuard(cache.Options{
		MemoryTTLSeconds: cfg.Cache.MemoryTTLSeconds,
	})
	limiter := middleware.NewRateLimiter(cfg.RateLimit.QPS, cfg.RateLimit.Burst)
	gameMap := dao.NewGameMap(cfg.Games)

	zap.L().Info("client-api start",
		zap.Int("port", cfg.ServerPort),
		zap.String("datacenter", cfg.DatacenterGrpc),
		zap.String("lottery", cfg.LotteryGrpc),
		zap.Strings("elastic", cfg.Elastic.Host),
		zap.Int("memoryTTL", cfg.Cache.MemoryTTLSeconds),
		zap.Int("rateQPS", cfg.RateLimit.QPS),
		zap.Int("gameMapSize", len(cfg.Games)))

	r := controller.NewRouter(controller.RouterDeps{
		DC:             dc,
		Lottery:        lottery,
		Guard:          guard,
		Limiter:        limiter,
		GameMap:        gameMap,
		InternalAPIKey: cfg.InternalAPIKey,
	})
	if err := r.Run(fmt.Sprintf(":%d", cfg.ServerPort)); err != nil {
		zap.L().Fatal("HTTP 启动失败", zap.Error(err))
	}
}

func main() {
	common.InitZapLogger()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%v\n", err)
	}
}
