# Redis 配置首次初始化

2026-11-03 从 `web-api/config.yaml` 指向的 Redis DB 0 导出全部 `/fg*` 配置，共 147 个 key，现场均为 string 类型。原始 key 和字符串值直接保存在 `web-api/dao/redis_defaults_data.go` 的 `defaultRedisConfigs` map 中，编译进 web-api，无需单独的 JSON 文件。

web-api 连接 Redis 后、加载内存配置前执行一次初始化：

- 标记 `fg_redis_configs_initialized` 已存在：完全跳过初始化。
- 标记不存在：将快照中的配置通过 `SET NX` 写入，仅补充缺失 key，保留现有值。
- 全部写入成功后保存永久标记。配置同样永久保存，不设过期时间。
- 通过 Lua 脚本串行处理多实例同时启动；初始化报错会中止启动，未完成时不写入成功标记，下一次启动可继续补齐。

原先启动时从 MySQL `pool_config` 无条件覆盖 Redis 的步骤已替换为上述初始化。后台编辑水池配置时仍同时保存 MySQL 和 Redis。清空 Redis 后首次启动使用内嵌快照，不再自动从 MySQL 恢复后来修改的配置；清空前应备份当前 Redis。

初始化后，即使手动删除某个配置，后续启动也不会再补回。如需再次补齐快照中的缺失配置，删除 `fg_redis_configs_initialized` 后重启 web-api；现有配置仍会保留。更新快照需要重新编译。

本次现场 Redis 仅用于读取和导出，未启动服务或向现场 Redis 写入标记。自动化测试使用 miniredis 验证首次导入、重复启动跳过、保留既有配置、永久保存与并发启动。
