# client-record-restored

这个目录是根据 `record.chichengwld.com` 暴露出来的 webpack source map 还原出的一个可运行 Vue 2 前端项目。

## 当前包含内容

- `src/components` 下的恢复组件
- `src/` 下的应用入口文件
- `src/vuex` 下的状态管理文件
- `src/assets/js` 下的工具脚本
- `src/components/Common/lang` 下的语言包
- `src/assets/css` 下的样式文件
- 一套兼容当前恢复代码的 webpack 构建配置

## 安装依赖

```bash
npm install
```

## 启动开发环境

```bash
npm run dev
```

## 本地调试

Windows 下可以直接运行：

```bat
start-local.bat
```

它会使用当前默认配置启动本地调试，并输出一个可直接访问的地址，例如：

```text
http://127.0.0.1:8080/fruit_record?token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiIsImp0aSI6IjRmMWcyM2ExMmFhYmxpenppIn0.eyJpc3MiOiJodHRwczpcL1wvaDUuY3ltc3kuY25cLyIsImF1ZCI6Imh0dHBzOlwvXC9oNS5jeW1zeS5jblwvIiwianRpIjoiNGYxZzIzYTEyYWFibGl6emkiLCJpYXQiOjE3NzY3NzIyNTcsIm5iZiI6MTc3Njc3MjI1NywiZXhwIjoxNzc3MDMxNDU3LCJ1c2VyX2luZm8iOiJ7XCJ1c2VyX2lkXCI6NDc3NTM5MyxcInVzZXJfbmFtZVwiOlwidWF0X3Rlc3Q0RkdEWjRBODAxMDM3MjVcIixcImdhbWVfaWRcIjpcIjcwMThcIixcImFnZW50X2lkXCI6Nzc0NyxcInRvcF9hZ2VudF91aWRcIjoxNTkzLFwiY3Vycm5lY3lpZFwiOlwiLVwifSJ9._e-fge9StXFmDamFYcGqwjLRAEIjKdrscenIfIVQ5K8&language=zh-cn&timezone=8
```

这个地址和 `src/components/Login.vue` 的逻辑一致：

- 从 query 中读取 `token`
- 从 query 中读取 `game_id`
- 从 query 中读取 `language`
- 从 query 中读取 `timezone`
- 写入 `Lockr` 后跳转到 `/slot_record`

## 打包

```bash
npm run build
```

## 环境变量

如需覆盖默认配置，可以参考 `.env.example`：

- `HOST`：接口基础地址，对应 `axios.defaults.baseURL`
- `DY_HOST`：部分棋盘详情/回放接口使用的地址
- `IMG_HOST`：挂到 `window.imgUrl` 的图片基础地址
- `ROUTER_BASE`：路由和 webpack 的公共基础路径
- `PORT`：本地开发端口

当前默认值指向：

```text
https://h5.chichengwld.com/
```

默认开发端口为：

```text
8081
```

## 当前说明

- 项目现在已经可以正常构建。
- 当前主要剩余的是 bundle 体积告警，不影响运行。
- 部分源码是从 source map 恢复的，原始工程里的构建配置、锁文件、历史依赖版本并没有完整恢复。
- `chess_detail.vue` 的原始恢复源码损坏较严重，目前替换成了一个稳定可运行的通用详情版本，优先保证项目可启动、可调试。
- 关闭按钮图片资源缺失的地方，已经改成文字按钮或兼容写法处理。
