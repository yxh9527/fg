package esindex

// Prefix 索引前缀。需要换前缀时直接改这里，例如 "fg_"。
const Prefix = "pp_"

// Name 拼接完整索引名：Prefix + suffix（suffix 不含前缀，如 gp_settlement）。
func Name(suffix string) string {
	return Prefix + suffix
}

// 索引后缀常量（不含前缀）
const (
	SuffixSettlement         = "gp_settlement"       // 注单
	SuffixFlowingWater       = "gp_flowing_water"    // 账变流水
	SuffixPoolRecordLog      = "pool_record_log"     // 水池日志
	SuffixDataAnalysis       = "data_analysis"       // 日分析
	SuffixDataAnalysisRange  = "data_analysis_range" // 区间分析
	SuffixUserController     = "user_controller"     // 用户调控
	SuffixGameStates         = "game_states"         // 断线状态（遗留）
	SuffixFlowingWaterLegacy = "flowing_water"       // crontab 旧名，仅清理用
)

func Settlement() string         { return Name(SuffixSettlement) }
func FlowingWater() string       { return Name(SuffixFlowingWater) }
func PoolRecordLog() string      { return Name(SuffixPoolRecordLog) }
func DataAnalysis() string       { return Name(SuffixDataAnalysis) }
func DataAnalysisRange() string  { return Name(SuffixDataAnalysisRange) }
func UserController() string     { return Name(SuffixUserController) }
func GameStates() string         { return Name(SuffixGameStates) }
func FlowingWaterLegacy() string { return Name(SuffixFlowingWaterLegacy) }
