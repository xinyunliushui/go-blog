## go-blog

## 简介
- 博客以及管理后台，前后端分离，此项目为后端服务。其中包含用户管理、角色管理、资源管理以及博客列表管理。
- 后端Go包含了gin、gorm、go-jwt、MySQL、RabbitMQ、ElasticSearch和ClickHouse等的使用。其中MQ、ES、CH的使用仅为了学习。
- 管理后台权限采用RBAC方案。用户、角色、资源三张主表以及用户角色关联表、角色资源关联表。
- 对应的博客和管理平台前端项目[web-blog](https://github.com/xinyunliushui/web-blog)。

## 项目结构
项目目录结构参照[标准Go项目布局](https://github.com/golang-standards/project-layout/blob/master/README_zh.md)
```text
go-blog/
├── main.go                          # 入口
├── go.mod
├── go.sum
├── README.md
├── .gitignore
├── public_key.pem                   # RSA 公钥
├── private_key.pem                  # RSA 私钥
├── migrations/                      # MySQL 结构变更脚本（goose）
└── internal/
    ├── config/                      # 配置加载；config.yml 与 dev/test/prod 环境覆盖
    ├── common/                      # 公共处理：MySQL 初始化、traceId、日志、校验器
    ├── routes/                      # 路由总入口与各模块路由
    ├── middleware/                  # gin相关中间件JWT、CORS、限流、Trace
    ├── controller/                  # 控制层；处理HTTP请求
    ├── service/                     # 业务逻辑层；RabbitMQ 消费、Blog MQ Compensation 定时重试
    ├── repository/                  # 数据访问层，负责与数据源交互
    ├── model/                       # GORM模型层
    ├── dto/                         # 数据传输对象；用于定义接口响应结构
    ├── vo/                          # 视图对象层；用于定义接口入参结构
    ├── response/                    # 统一JSON响应
    ├── rabbitmq/                    # AMQP 连接、声明队列、发布与消费封装
    ├── elasticsearch/               # ES 客户端初始化
    ├── clickhouse/                  # ClickHouse 连接与迁移
    ├── utils/                       # 工具函数；RSA、BCrypt、JSON、环境变量、ES 高亮等
```

## 启动前依赖准备
以下环境以及预设好对应的`database`，用户名和密码记得更换
- `MySQL` 需要前置创建database，名称是go_blog
``` shell
# 注意数据库字符集格式和项目配置保持一致
```
- `RabbitMQ` 需要前置创建vhost，名称是go_blog，并给用户授权
```shell
# 本地docker示例
# 创建vhost go_blog
rabbitmqctl add_vhost go_blog
# 查看vhost go_blog的用户列表
rabbitmqctl list_permissions -p go_blog
# vhost授权给用户admin
rabbitmqctl set_permissions -p go_blog admin "." "." ".*"
```
- `Elasticsearch` 需要前置安装好ik_max_word中文分词器，优化博客文章存储
```shell
# 本地docker示例
# 安装对应版本的分词器
./bin/elasticsearch-plugin install https://release.infinilabs.com/analysis-ik/stable/elasticsearch-analysis-ik-<版本号>.zip
# 查看配置的插件
./bin/elasticsearch-plugin list
```
- `ClickHouse` 需要前置创建database，名称是go_blog
- 本地开发
```shell
# 1、安装依赖
go mod tidy
# 2、Goose管理迁移，参照下方说明
# 3、本地启动
go run .\main.go
```

## Goose（MySQL 迁移）
- 迁移脚本目录：`migrations/`，文件命名 `YYYYMMDDxxxx_描述.sql`，使用 `-- +goose Up` / `-- +goose Down` 分段。
- **开发环境**（`APP_ENV=dev`，默认）：启动时由 GORM `AutoMigrate` 按模型补齐表结构，便于本地迭代。
- **测试 / 生产环境**：进程启动不执行 goose；发版或部署前需用 goose CLI 手动执行 `migrations/` 脚本。
``` shell
# 安装 goose CLI
go install github.com/pressly/goose/v3/cmd/goose@latest

# 在项目根目录执行（DSN 与 config 中 mysql 一致，按实际替换）
goose -dir migrations mysql "user:password@tcp(host:3306)/go_blog?charset=utf8mb4&parseTime=True&loc=Local" up

# 查看迁移状态 / 回滚上一版
goose -dir migrations mysql "<dsn>" status
goose -dir migrations mysql "<dsn>" down
```

## Gin 中间件
- `AuthMiddleware` — 登录、登出与 JWT 校验
- `RateLimitMiddleware` — 令牌桶限流
- `CORSMiddleware` — 跨域请求处理


## 方案简要说明
- 密码传输采用 RSA 非对称加密，入库使用 BCrypt 不可逆哈希。
- 日志使用 Zap 记录，配合 Lumberjack 做按大小/时间的文件切割。
- API 通过路径前缀 `/v1` 做版本隔离。
- 提供存活探针与就绪探针；MySQL、MQ、ES、CH 等依赖初始化失败不阻塞进程启动，由就绪探针反映不可用状态。
- 主键统一为 UUID 字符串，在 GORM `BeforeCreate` 钩子中生成。
- 监听退出信号，按 HTTP → 补偿重试 → MQ 主消费 / DLQ 消费顺序优雅停机并释放连接。
- MQ 推送或消费失败时写入 `blog_mq_compensation` 补偿表，后台定时重试（PUBLISH 补发 MQ，CONSUME 补写 ES/CH）。
- 消费失败且补偿落库失败时，消息经 Broker 死信交换机进入 `go_blog_dlq`，由 DLQ 消费者再次尝试落补偿表。

## 特殊说明
### MQ 推送 / 消费失败补偿

博客创建后异步投递 RabbitMQ（`go_blog_exchange` → `go_blog_queue`），消费端将文章同步至 ES 与 ClickHouse。任一环节失败不阻塞主流程，依赖本地表 `blog_mq_compensation` 做最终一致性补偿；补偿落库也失败时由死信队列兜底。

**PUBLISH**场景
- 创建文章后 `PublishMessage` 失败 | 定时任务重新向 MQ 发布消息

**CONSUME**场景
- 消费后同步 ES/CH 失败 | 写 CONSUME 补偿并 Ack，由定时任务按 `pending_mask` 补写
- 消费失败且补偿落库失败 | Nack 转入 `go_blog_dlq`，DLQ 消费者再次写补偿表

**写入规则**

- 同 `blog_id` + `task_type` 且`status`为「待处理」时更新写入，避免重复任务；CONSUME 的 `pending_mask`（ES=1、CH=2）做 ES、CH 数据一致性判断。
- CONSUME 失败且补偿落库成功 → Ack；落库失败 → Nack 进死信队列（非丢弃）。

**后台重试**（`RunBlogMQCompensationRetry`）

- 每 30 秒扫描一批（最多 10 条），最多重试 5 次。
- 成功 → `status=1`；超限 → `status=2`（已放弃），日志关键字 `[MQ_PUBLISH_DEAD]` / `[MQ_CONSUME_DEAD]`，需人工补投 MQ 或核对 ES/CH。
- DLQ 再次落补偿仍失败 → `[MQ_DLQ_DEAD]`，需人工核对 ES/CH。


## 相关三方依赖
- [`Gin`](https://github.com/gin-gonic/gin) — HTTP API 框架
- `MySQL` — 业务主库（用户、角色、资源、博客等）
- [`gin-jwt`](https://github.com/appleboy/gin-jwt) — JWT 登录认证与 Token 刷新
- [`Gorm`](https://github.com/go-gorm/gorm) — ORM，支持关联、事务与表迁移
- [`Validator`](https://github.com/go-playground/validator) — 请求入参校验（v10）
- [`Lumberjack`](https://github.com/natefinch/lumberjack) — 日志文件轮转（大小、份数、保留期、压缩）
- [`Viper`](https://github.com/spf13/viper) — 配置加载与管理
- [`GoFunk`](https://github.com/thoas/go-funk) — Slice 等集合工具函数，[文档](https://pkg.go.dev/github.com/thoas/go-funk#pkg-index)
- [`streadway/amqp`](https://github.com/streadway/amqp) — RabbitMQ AMQP 客户端，连接管理、队列声明、消息发布与消费
- [`go-elasticsearch`](https://github.com/elastic/go-elasticsearch) — Elasticsearch 官方 Go 客户端（v8），博客索引、全文搜索与 ES 对账
- [`Gorm ClickHouse Driver`](https://github.com/go-gorm/clickhouse) — ClickHouse 的 GORM 驱动（底层 [`clickhouse-go/v2`](https://github.com/ClickHouse/clickhouse-go)），博客异步同步与 OLAP 存储
