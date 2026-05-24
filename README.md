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

## 使用Docker快速启动测试环境，体验本项目
测试环境使用 Docker Compose 一键启动应用及全部依赖（MySQL、RabbitMQ、Elasticsearch、ClickHouse）。相关文件位于 `docker/` 目录：

```text
docker/
├── docker-compose.yml          # 服务编排（含健康检查、初始化 Job）
├── Dockerfile                  # go-blog 多阶段构建
├── Dockerfile.elasticsearch    # Elasticsearch + IK 分词
├── .env.example                # 环境变量模板（复制为 .env）
├── daemon.json.example         # Docker Desktop 国内 registry-mirrors
├── .gitignore
└── init/                       # 依赖初始化脚本
    ├── mysql/
    ├── rabbitmq/
    ├── clickhouse/
    └── elasticsearch/
```

### 启动前准备

1. 安装并启动 [Docker Desktop](https://www.docker.com/products/docker-desktop/)，确认 Engine 处于 running 状态。
2. 复制并修改环境变量模板：`copy docker\.env.example docker\.env`。密码、vhost、镜像版本等统一在 `.env` 配置；应用连接参数通过环境变量注入（优先级高于 `internal/config/config.test.yaml`）。
3. （推荐）Docker Desktop → Settings → Docker Engine，合并 `docker/daemon.json.example` 中的 `registry-mirrors`。
4. 基础配置仍使用 `internal/config/config.yml` 与 `internal/config/config.test.yaml`（队列名、交换机名等非敏感项）。

### 启动与停止

在项目根目录或 `docker` 目录下执行均可，以下以 `docker` 目录为例：

```shell
# 进入目标文件夹
cd docker

# 构建并后台启动全部服务（首次会构建镜像，耗时较长）
docker compose up -d --build

# 查看容器状态
docker compose ps

# 查看应用日志
docker compose logs -f go-blog

# 停止服务（保留数据卷）
docker compose down

# 彻底结束当前项目并清理相关镜像（不会清空数据卷和缓存）
docker compose down --rmi all

# 停止并删除数据卷（清空数据库等持久化数据）
docker compose down -v
```

也可在 Docker Desktop 中打开 `docker/docker-compose.yml` 所在目录，通过 Compose 的 **Build** / **Up** 启动。

### 访问地址

| 服务 | 地址 |
|------|------|
| API | http://localhost:8080/api |
| 存活探针 | http://localhost:8080/health |
| 就绪探针 | http://localhost:8080/ready |
| RabbitMQ 管理台 | http://localhost:15672（admin / 123456） |
| Elasticsearch | http://localhost:9200 |
| ClickHouse HTTP | http://localhost:8123 |

就绪探针 `/ready` 返回 200 表示 MySQL、RabbitMQ、Elasticsearch、ClickHouse 均已连通。

### 说明

- `docker/.env` 中的 `RABBITMQ_VHOST`、`MYSQL_DATABASE`、`CLICKHOUSE_DB` 等会同步用于依赖初始化（`rabbitmq-init`、`clickhouse-init` 等）及应用连接，修改后执行 `docker compose up -d` 重建相关容器。
- 容器内 `APP_ENV=test`，MySQL 表结构由 GORM `AutoMigrate` 自动补齐；若有 `migrations/` 增量脚本，可在宿主机用 goose CLI 连接 `127.0.0.1:3306` 手动执行。
- 修改 `internal/config` 下非敏感配置后，重启 `go-blog` 容器即可生效（配置目录已挂载进容器）。
- 所有镜像拉取与构建均默认使用国内源（见 `docker/.env.example`）；第三方镜像版本集中在 `.env` 的 `*_IMAGE_TAG` 字段，升级时只需改一处。
- MySQL 若报 `Could not open the mysql.plugin table`，多为旧数据卷与镜像版本不兼容，执行 `docker compose down -v` 清空数据卷后重新 `up`。
- RabbitMQ 若长时间 `unhealthy`，先 `docker compose logs go-blog-rabbitmq` 查看日志；仍失败时执行 `docker compose down -v` 重建 `rabbitmq_data` 后重试。
- 首次启动时 Elasticsearch 需构建 IK 分词插件镜像，全部依赖 healthcheck 通过后应用才会启动，请耐心等待。
- 根目录 `.gitattributes` 规定 `docker/init/**` 下脚本统一使用 LF（Unix）换行。这些脚本会在 Linux 容器内执行；Windows 上若检出为 CRLF，可能导致 `sh` 报错（如 `not found`、`bad interpreter`）。Mac 默认即为 LF，一般无此问题；该配置主要保证跨平台协作时脚本在容器内可正常执行。


## 本地开发以及前置依赖安装
### 前置依赖
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
### 本地开发
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

