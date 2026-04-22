package main

import (
	"app/config"
	"app/tables"
	"fmt"
	"os"
	"web-api-ex/common"
	. "web-api-ex/common"
	"web-api-ex/controller"
	"web-api-ex/dao"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	rootCmd = &cobra.Command{
		Use:  "web-api-ex",
		Long: "web-api-ex",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launch the server",
		Run:   run,
	}
)

func init() {
	runCmd.Flags().String("config", "./config.yaml", "指定配置文件，默认使用当前目录下的config.yaml")
	rootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) {
	path, err := cmd.Flags().GetString("config")
	if err != nil {
		return
	}
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		zap.L().Fatal("读取基础配置失败", zap.Any("error", err))
	}
	rc := &config.RunConfig{}
	err = yaml.Unmarshal(yamlFile, rc)
	if err != nil {
		zap.L().Fatal("解析基础配置异常", zap.Any("error", err))
	}
	//初始化etcd
	dao.NewRedisDao(rc.Redis.Host, rc.Redis.User, rc.Redis.Pwd)
	//初始化配置信息
	dao.ConfigsInit()
	//初始化es
	if err := dao.InitES(rc); err != nil {
		panic(err)
	}
	//初始化mysql
	if err := dao.InitDB(rc); err != nil {
		panic(err)
	}
	//初始化数据库
	tables.InitMysqlDb(dao.Mysql().Manager, dao.Mysql().Player)
	//初始化权限策略
	common.NewUrlPloy()
	r := controller.NewRouter()
	zap.L().Info("服务器启动成功")
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
