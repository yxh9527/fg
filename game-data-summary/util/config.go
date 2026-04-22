package util

import (
	"app/config"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// 解析配置
func ParseConfig(key string, value string) {
	arr := strings.Split(key, "/")
	if len(arr) < 2 || value == "" {
		zap.L().Error("配置数据异常", zap.Any("key", key), zap.Any("data", value))
	} else {
		if arr[1] == "config" {
			switch arr[2] {
			// /config/system
			case "system":
				tmp := &config.SystemConfig{}
				if err := jsoniter.UnmarshalFromString(value, tmp); err == nil {
					config.CfgIns.SetSystemConfig(tmp)
				} else {
					zap.L().Error("加载系统配置失败", zap.Any("err", err), zap.Any("value", value))
				}
			case "currency":
				tmp := config.CfgIns.Currency
				if err := jsoniter.UnmarshalFromString(value, tmp); err == nil {
					config.CfgIns.SetCurrency(tmp)
				} else {
					zap.L().Error("加载系统配置失败", zap.Any("err", err), zap.Any("value", value))
				}
			// /config/pool/{symbol}
			case "pool":
				tmp := &config.Pool{}
				if err := jsoniter.UnmarshalFromString(value, tmp); err == nil {
					zap.L().Debug("加载pool配置文件成功", zap.Any("data", tmp))
					config.CfgIns.SetDefaultPool(tmp.Symbol, tmp)
				} else {
					zap.L().Error("加载pool配置失败", zap.Any("err", err))
				}
			// /config/ctrl/{symbol}
			case "ctrl":
				tmp := &config.AwardConfig{}
				if err := jsoniter.UnmarshalFromString(value, tmp); err == nil {
					if arr[3] != "default" && tmp.GameId == 0 {
						zap.L().Debug("ctrl 配置异常", zap.Any("data", tmp))
						return
					}
					config.CfgIns.SetCtrl(tmp.Symbol, tmp)
				} else {
					zap.L().Error("加载ctrl配置失败", zap.Any("err", err))
				}
			// /config/autoCtrl
			case "autoCtrl":
				tmp := &config.AutoCtrlMgr{}
				if err := jsoniter.UnmarshalFromString(value, tmp); err == nil {
					config.CfgIns.SetAutoCtrl(tmp)
				} else {
					zap.L().Error("加载autoCtrl配置失败", zap.Any("err", err))
				}
			}
		}
		if arr[1] == "agent" {
			// /agent/{agentId}/pool/{symbol}
			if arr[3] == "pool" {
				tmp := &config.Pool{}
				if err := jsoniter.UnmarshalFromString(value, tmp); err == nil {
					config.CfgIns.SetAgentPool(arr[2], tmp)
				} else {
					zap.L().Error("加载代理pool配置失败", zap.Any("err", err))
				}
			}
		}
	}
}

// 初始化基础配置
func InitBaseConfig(path string) *config.RunConfig {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		zap.L().Fatal("读取基础配置失败", zap.Any("error", err))
	}
	config := &config.RunConfig{}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		zap.L().Fatal("解析基础配置异常", zap.Any("error", err))
	}
	zap.L().Debug("=====>", zap.Any("config", config))
	return config
}
