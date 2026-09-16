# Elasticsearch 索引清单

索引名统一由 `app/esindex` 生成。

## 改前缀

只改代码常量即可（不走配置文件）：

```go
// app/esindex/esindex.go
const Prefix = "pp_" // 改成 "fg_"
```

## 代码引用

| 方法 | 后缀 | Prefix=`pp_` | Prefix=`fg_` |
| --- | --- | --- | --- |
| `esindex.Settlement()` | `gp_settlement` | `pp_gp_settlement` | `fg_gp_settlement` |
| `esindex.FlowingWater()` | `gp_flowing_water` | `pp_gp_flowing_water` | `fg_gp_flowing_water` |
| `esindex.PoolRecordLog()` | `pool_record_log` | `pp_pool_record_log` | `fg_pool_record_log` |
| `esindex.DataAnalysis()` | `data_analysis` | `pp_data_analysis` | `fg_data_analysis` |
| `esindex.DataAnalysisRange()` | `data_analysis_range` | `pp_data_analysis_range` | `fg_data_analysis_range` |
| `esindex.UserController()` | `user_controller` | `pp_user_controller` | `fg_user_controller` |
| `esindex.GameStates()` | `game_states` | `pp_game_states` | `fg_game_states` |
| `esindex.FlowingWaterLegacy()` | `flowing_water` | `pp_flowing_water` | `fg_flowing_water` |

## 注意

1. 所有服务共用 `app/esindex.Prefix`，改一处即可全局生效。
2. 改前缀不会自动迁移旧索引数据。
3. 业务代码禁止写死 `"pp_xxx"`，一律用 `app/esindex`。
