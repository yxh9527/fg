# client-api

面向**客户端**的 HTTP 服务（新服 `clientApi` 记录页 / 登录资料等）。

## 职责边界（重要）

| 做 | 不做 |
| --- | --- |
| 鉴权 / 登录资料（转 data-center） | **不做** slots/fruits 结算 |
| 注单列表 / 详情（直连 ES） | **不做** GameStorage 写入 |
| 客户端余额展示（转 lottery.GetBalance） | **不做** 游戏服结算中转 |

游戏服 `SlotsRpcProd` / `StorageRpcProd` / `CommonRpcProd` 应 **直连 lottery gRPC**，不要打到本服务。

## 防御

1. **防重入**：`singleflight`
2. **内存缓存**：TTL 内相同参数直接返回（不写 Redis）
3. **IP 限流**
4. **启动自检**：ES + data-center + lottery

## 启动

```bash
go run . run --config ./config.yaml
```

## 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/client/v1/auth/authenticate` | 登录鉴权 |
| POST | `/api/client/v1/auth/validate` | token 绑定校验 |
| GET | `/api/client/v1/auth/login-data?userId=` | 登录资料 |
| GET | `/api/client/v1/balance?userId=` | 余额展示 |
| POST | `/api/client/v1/record/authenticate` | 记录页鉴权 |
| GET | `/api/client/v1/record/list` | 注单列表 |
| GET | `/api/client/v1/record/detail` | 注单详情 |
| GET | `/api/client/v1/internal/record/*` | S2S 查单（`X-Internal-Key`） |
| GET | `/api/client/v1/internal/statement/list` | S2S 流水（ES `pp_gp_flowing_water`） |
| GET | `/api/client/v1/internal/balance` | S2S 余额展示 |

## 环境变量（fg-game-server）

```env
# 仅 ClientApi / 记录页
FG_CLIENT_API_BASE_URL=http://127.0.0.1:10092
FG_CLIENT_API_INTERNAL_KEY=change-me-client-api-internal

# 游戏结算 / 余额权威：直连 lottery
FG_LOTTERY_GRPC=172.21.211.219:10080
FG_LOTTERY_PROTO_DIR=D:/project/fg/micro_service/proto
```
