## go-blog

## 简介
- 博客以及管理后台，前后端分离，此项目为后端服务。
- 后端Go包含了gin、gorm、jwt、MySQL、RabbitMQ、ElasticSearch和ClickHouse等的使用。其中MQ、ES、CH的使用仅为了学习。
- 管理后台权限管理采用RBAC方案。
- 对应的博客和管理平台前端项目[web-blog](https://github.com/xinyunliushui/web-blog)。

## 项目结构
项目目录结构参照[标准Go项目布局](https://github.com/golang-standards/project-layout/blob/master/README_zh.md)

```text
go-blog/
├── main.go                          # 入口：配置加载、组件初始化、HTTP 服务、MQ 消费与 Outbox 重试协程
├── go.mod
├── go.sum
├── README.md
├── .gitignore
├── public_key.pem                   # RSA 公钥
├── private_key.pem                  # RSA 私钥
└── internal/
    ├── config/                      # 项目相关配置
    ├── common/                      # 公共处理：MySQL 初始化与迁移、日志、校验器、默认管理员
    ├── routes/                      # 路由总入口与各模块路由
    ├── middleware/                  # gin相关中间件JWT、CORS、限流
    ├── controller/                  # 控制层；处理HTTP请求
    ├── service/                     # 业务逻辑层；RabbitMQ 消费、Blog_MQ_Outbox定时重试
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


## 启动前准备
以下环境以及预设好对应的`database`，用户名和密码记得切换
- `MySQL` 需要前置创建database，名称是go_blog
- `RabbitMQ` 需要前置创建vhost，名称是go_blog
- `Elasticsearch` 需要前置安装好ik_max_word中文分词器
- `ClickHouse` 需要前置创建database，名称是go_blog


## 技术栈
- `Gin` 一个类似于martini但拥有更好性能的API框架, 由于使用了httprouter, 速度提高了近40倍
- `MySQL` 采用的是MySql数据库
- `Jwt` 使用JWT轻量级认证, 并提供活跃用户Token刷新功能
- `Gorm` 采用Gorm 2.0版本开发, 包含一对多、多对多、事务等操作
- `Validator` 使用validator v10做参数校验, 严密校验前端传入参数
- `Lumberjack` 设置日志文件大小、保存数量、保存时间和压缩等
- `Viper` Go应用程序的完整配置解决方案, 支持配置热更新
- `GoFunk` 包含大量的Slice操作方法的工具包

## gin中间件
- `AuthMiddleware` 权限认证中间件 -- 处理登录、登出、无状态token校验
- `RateLimitMiddleware` 基于令牌桶的限流中间件 -- 限制用户的请求次数
- `CORSMiddleware` -- 跨域中间件 -- 解决跨域问题


## 其他
- 密码的传输使用非对称加密、入库使用不可逆加密
- 日志管理使用Zap配合Lumberjack（由于zap不具备日志切割功能, 使用lumberjack配合）

## TODO
- MQ消费失败后目前直接丢弃了，需要考虑补偿方案
- ES和CH入库非原子操作，需要两边数据一致性问题
- 项目的log数据可以尝试接入ES，实现项目的日志的分析
- 博客的运营数据写入CH，让管理后台具有数据洞察能力


## MQ发送失败时落库并定时重试，实现说明（使用AI实现）:
`controller` 写博客成功后调用 `rabbitmq` 发布消息；`ConsumeRabbitMQ` 消费后通过 `repository` 写入 ES 与 ClickHouse；若发布失败则由 `repository` 写入 Outbox 表，`RunBlogMQOutboxRetry` 定时读表并补发

#### 表模型 internal/model/blog_mq_outbox_model.go
- 表名：blog_mq_outbox
- 字段：blog_id、完整 Blog JSON（payload）、status（0 待投递 / 1 成功 / 2 放弃）、retry_count、last_error

#### 仓储 internal/repository/mq_outbox_repository.go
- EnqueueBlogPublish：MQ 首次失败后写入一行待投递记录
- ListPendingForRetry：拉取 status=待投递 的记录
- MarkSent / MarkRetry：成功改「已发送」，失败递增重试次数，超过上限改「已放弃」

#### 后台重试 internal/service/blog_mq_outbox_retry_service.go
- 每 30s 扫一批（最多 50 条），反序列化 payload 后再次 PublishMessage
- 单条最多 15 次失败后标记 DEAD，并打 [MQ_OUTBOX_DEAD_LETTER] 日志便于告警规则抓取
- 任意异常统一带 [MQ_OUTBOX_ALERT]，便于监控检索
- 常量：blogOutboxRetryInterval、blogOutboxMaxRetries、blogOutboxBatchSize（可按需改成配置项）。

#### 创建博客 internal/controller/blog_controller.go
- MySQL Create 成功后照常 PublishMessage
- 失败：打 [MQ_OUTBOX_ALERT] → 写入补偿表 → 返回成功文案「…将自动重试」并带上 blogId
- 补偿表写入也失败：返回业务失败，提示人工介入（避免误以为已排队）

#### 进程生命周期 main.go
- 启动独立 goroutine：RunBlogMQOutboxRetry
- 优雅退出：先 stopOutbox 并 Wait，再停 MQ 消费者并 CloseRabbitMQ，避免关闭连接后仍在 Publish。
