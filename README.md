## go-blog

## 简介
- 博客以及管理后台，前后端分离，此项目为后端服务。
- 后端Go包含了gin、gorm、jwt、MySQL、RabbitMQ、ElasticSearch和ClickHouse等的使用。
- 管理后台权限管理采用RBAC方案。
- 对应的博客和管理平台[web-blog](https://github.com/xinyunliushui/web-blog)。

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
- 博客在创建的时候就写入ES和CH，博客编辑后未更新ES和CH中的数据
- log数据可以尝试接入ES，实现日志的分析
- 博客的运营数据写入CH，让管理后台具有数据洞察能力
