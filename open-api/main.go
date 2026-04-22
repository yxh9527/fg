package main

import (
	"app/config"
	"fmt"
	. "open-api/common"
	"open-api/controller"
	"open-api/dao"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	rootCmd = &cobra.Command{
		Use:  "open-api",
		Long: "open-api",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "启动open-api",
		Long:  "启动open-api",
		Run:   run,
	}
)

var (
	RunConfigPath string
)

func init() {
	runCmd.Flags().StringVar(&RunConfigPath, "config", "./config.yaml", "指定配置文件，默认使用当前目录下的config.yaml")
	rootCmd.AddCommand(runCmd)
}

// 初始化基础配置
func InitBaseConfig() *config.RunConfig {
	yamlFile, err := os.ReadFile(RunConfigPath)
	if err != nil {
		zap.L().Fatal("读取基础配置失败", zap.Any("error", err))
	}
	c := &config.RunConfig{}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		zap.L().Fatal("解析基础配置异常", zap.Any("error", err))
	}
	return c
}

func run(cmd *cobra.Command, args []string) {
	rc := InitBaseConfig()
	dao.InitRedis(rc)
	dao.ConfigsInit()
	dao.LoadConfig()
	if err := dao.InitES(rc); err != nil {
		panic(err)
	}
	if err := dao.InitDB(rc); err != nil {
		panic(err)
	}
	//初始化代理缓存
	dao.InitAgentMgr()
	//动态域名加载
	InitApiConfigMgr()
	//加载游戏配置
	dao.InitGameCacheMgr()
	zap.L().Info("Server ", zap.String("name", "open-api"))
	zap.L().Info("Server start ok")
	r := controller.NewRouter()
	if err := r.Run(fmt.Sprintf(":%d", rc.ServerPort)); err != nil {
		zap.L().Fatal("HTTP Server启动失败", zap.Error(err))
	}
}

func main() {
	// 初始化日志库
	InitZapLogger()

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%v", err)
	}
}
