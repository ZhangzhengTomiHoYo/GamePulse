# GamePulse 项目文档（Bluebell）

文档版本：v1.0  
整理时间：2026-03-16  
项目路径：`e:\BaiduSyncdisk\Golang\GamePulse\Project\GamePulse`

## 1. 项目概览

`GamePulse`（代码模块名 `bluebell`）是一个基于 Go + Gin 的社区帖子系统后端项目，提供用户注册登录、社区查询、帖子发布、帖子列表查询（按时间/分数排序）和投票能力，并集成了 Swagger 文档、pprof、结构化日志、JWT 鉴权、Redis 缓存与排序。

项目核心目标是实现一个典型的论坛/社区后端分层架构：

- `controller` 负责 HTTP 参数校验与响应
- `logic` 负责业务编排
- `dao/mysql` 负责持久化
- `dao/redis` 负责缓存、排序和投票计分

## 2. 功能清单

- 用户注册：`POST /api/v1/signup`
- 用户登录：`POST /api/v1/login`
- 社区列表：`GET /api/v1/community`（需登录）
- 社区详情：`GET /api/v1/community/:id`（需登录）
- 创建帖子：`POST /api/v1/post`（需登录）
- 帖子详情：`GET /api/v1/post/:id`（需登录）
- 帖子列表（旧版）：`GET /api/v1/posts`（需登录）
- 帖子列表（新版，支持排序/按社区筛选）：`GET /api/v1/posts2`（需登录）
- 帖子投票：`POST /api/v1/vote`（需登录）
- 其他：`GET /ping`、`GET /swagger/*any`、`GET /`（前端静态入口）

## 3. 技术栈与依赖

- 语言与运行时：Go `1.25`
- Web 框架：`gin-gonic/gin`
- 配置：`spf13/viper`
- 日志：`zap` + `lumberjack`（按大小滚动）
- 数据库：MySQL（`sqlx` + `go-sql-driver/mysql`）
- 缓存：Redis（`go-redis`）
- 鉴权：JWT（`dgrijalva/jwt-go`）
- ID 生成：Snowflake（`bwmarrin/snowflake`）
- 文档：Swagger（`swaggo/gin-swagger`）
- 运维调试：`gin-contrib/pprof`
- 限流：`juju/ratelimit`

## 4. 项目目录说明

- `main.go`：应用启动入口，初始化配置、日志、MySQL、Redis、Snowflake、翻译器和路由
- `conf/config.yaml`：项目配置
- `routes/`：路由注册与全局中间件
- `controllers/`：HTTP 接口层
- `logic/`：业务逻辑层
- `dao/mysql/`：MySQL 数据访问
- `dao/redis/`：Redis 数据访问与投票排序
- `models/`：请求/响应模型与数据库结构定义
- `middlewares/`：JWT 鉴权、限流中间件
- `pkg/`：通用组件（JWT、Snowflake）
- `setting/`：配置加载
- `logger/`：日志初始化和 Gin 日志/恢复中间件
- `docs/`：Swagger 产物
- `templates/` + `assets/`：前端静态资源
- `Dockerfile`、`docker-compose.yml`、`wait-for.sh`：容器化与启动编排

## 5. 架构设计

### 5.1 分层架构

请求链路：

`Gin Router -> Middleware -> Controller -> Logic -> DAO(MySQL/Redis) -> Response`

职责边界：

- Controller：参数绑定、基础校验、统一返回格式
- Logic：组合查询、业务规则、异常映射
- DAO：单一数据源读写

### 5.2 中间件

- 日志中间件：记录请求路径、状态码、耗时、UA、IP
- Recovery 中间件：捕获 panic，防止进程崩溃
- 全局限流：令牌桶，容量 100，每秒补充（全局共享，不区分用户）
- JWT 鉴权：校验 `Authorization: Bearer <token>`，解析后将 `userID` 写入上下文

## 6. 配置说明（`conf/config.yaml`）

### 6.1 基础配置

- `name`: 项目名（`bluebell`）
- `mode`: 运行模式（`dev`/`release`）
- `port`: 服务端口（默认 `8080`）
- `version`: 版本号
- `start_time`: Snowflake 起始时间（格式 `YYYY-MM-DD`）
- `machine_id`: Snowflake 节点 ID

### 6.2 鉴权配置

- `auth.jwt_expire`: JWT 过期小时数（默认 `8760` 小时，约 1 年）

### 6.3 日志配置

- `log.level`: 日志级别
- `log.filename`: 日志文件名
- `log.max_size`: 单文件最大 MB
- `log.max_age`: 保留天数
- `log.max_backups`: 备份数

### 6.4 数据源配置

- MySQL：`host`、`port`、`user`、`password`、`dbname`、连接池参数
- Redis：`host`、`port`、`password`、`db`、`pool_size`

## 7. 数据模型与存储设计

### 7.1 MySQL 表

定义文件：`models/create_table.sql`

- `user`：用户信息（`user_id`、`username`、`password` 等）
- `community`：社区分类（`community_id`、`community_name`、`introduction`）
- `post`：帖子信息（`post_id`、`title`、`content`、`author_id`、`community_id`）

说明：

- 用户密码存储为 `md5(secret + password)` 形式（当前实现）
- `post_id`、`user_id` 为业务主键（Snowflake）

### 7.2 Redis Key 设计

键前缀：`bluebell:`

- `post:time`（ZSet）：帖子发布时间排序
- `post:score`（ZSet）：帖子热度分数排序
- `post:voted:{postID}`（ZSet）：某帖子下用户投票记录（member=userID，score=1/0/-1）
- `community:{communityID}`（Set）：某社区的帖子 ID 集合

社区排序查询时会构造临时 ZSet（TTL 60 秒），用于社区集合与排序集合交集计算。

## 8. API 说明

基础路径：`/api/v1`  
统一响应结构：

```json
{
  "code": 1000,
  "msg": "success",
  "data": {}
}
```

### 8.1 错误码

- `1000`：成功
- `1001`：参数错误
- `1002`：用户已存在
- `1003`：用户不存在
- `1004`：用户名或密码错误
- `1005`：服务繁忙
- `1006`：无效 Token
- `1007`：需要登录

### 8.2 无需登录接口

- `POST /signup`：注册
- `POST /login`：登录

### 8.3 需登录接口

请求头统一：

`Authorization: Bearer <token>`

- `GET /community`：社区列表
- `GET /community/:id`：社区详情
- `POST /post`：发帖
- `GET /post/:id`：帖子详情
- `GET /posts`：旧版分页列表
- `GET /posts2`：新版分页列表（`page`、`size`、`order`、`community_id`）
- `POST /vote`：投票（`post_id`，`direction` 为 `1/0/-1`）

## 9. 核心业务流程

### 9.1 注册流程

1. Controller 绑定 `username/password/re_password`
2. Logic 检查用户是否存在
3. Snowflake 生成用户 ID
4. DAO 写入 MySQL（密码加密后入库）

### 9.2 登录流程

1. MySQL 查询用户
2. 校验密码
3. 生成 JWT（包含 `user_id` 和 `username`）
4. 返回 token 给客户端

### 9.3 发帖流程

1. 鉴权中间件解析用户身份
2. 生成 `post_id`
3. 写入 MySQL
4. 写入 Redis 排序集合（时间/分数）并加入社区集合

### 9.4 帖子列表流程（`/posts2`）

1. 根据 `order` 决定 Redis 排序 key（时间或分数）
2. 读取分页帖子 ID（Redis）
3. 按 ID 顺序回源 MySQL 查询帖子详情
4. 批量查询投票统计（Redis pipeline）
5. 聚合作者、社区、投票数并返回

### 9.5 投票流程

1. 校验帖子发布时间是否超过 1 周
2. 读取用户历史投票值
3. 按 `diff * scorePerVote(432)` 更新帖子分数
4. 更新或删除用户投票记录（方向 0 表示取消）

## 10. 本地运行指南

### 10.1 开发环境

- Go `1.25`
- MySQL `8.0`
- Redis 最新版

### 10.2 启动步骤（推荐）

1. 启动依赖：
`docker compose up -d mysql redis`
2. 初始化数据库结构：
执行 `models/create_table.sql`
3. 根据运行环境修改 `conf/config.yaml`（本机运行通常将 `mysql/redis` 主机改为 `127.0.0.1`）
4. 启动应用：
`go run .`

访问地址：

- API：`http://localhost:8080`
- Swagger：`http://localhost:8080/swagger/index.html`
- pprof：`http://localhost:8080/debug/pprof`

### 10.3 容器化运行

项目提供 `Dockerfile` 与 `docker-compose.yml`，应用映射端口为 `8888:8080`。

## 11. 测试现状

执行时间：2026-03-16  
执行命令：`go test ./...`

结果摘要：

- `controllers` 包测试通过
- `dao/mysql` 包测试失败，原因是本地 `127.0.0.1:3306` 不可达（测试初始化阶段直接连接数据库）
- 其余包无测试文件或通过

结论：

- 当前项目具备基础测试，但依赖真实 MySQL，尚未做测试隔离

## 12. 已知问题与风险

- `init.sql` 仅创建数据库与 root 认证方式，不会创建业务表；首次运行仍需手动执行 `models/create_table.sql`
- `wait-for.sh` 中存在 `local all_ready=1`（在函数外使用 local），在部分 `/bin/sh` 环境可能报错
- `controllers/vote.go` 参数校验失败后有一个分支未 `return`，会继续执行后续逻辑
- 密码算法使用 MD5 + 固定盐，安全性不足，不适合生产
- JWT 密钥写死在代码中，建议改为配置或密钥管理服务
- 仓库包含大体积日志文件（`web_app.log`、`web_app-*.log`），建议加入 `.gitignore` 并清理

## 13. 优化建议（按优先级）

1. 修复启动与稳定性问题：完善 `init.sql` 表结构初始化，修复 `wait-for.sh` 兼容性
2. 修复业务逻辑瑕疵：补充 `PostVoteController` 参数错误后的 `return`
3. 增强安全性：升级密码哈希（`bcrypt/argon2`），将 JWT 密钥外置化
4. 测试改造：为 DAO 增加可替换数据源或测试容器，减少对本地环境耦合
5. 工程治理：补全 README、清理日志产物、增加 CI（`go test` + `go vet` + lint）

## 14. 维护建议

- 新增接口时保持 `controller -> logic -> dao` 分层，不跨层调用
- Redis key 统一通过 `getRedisKey` 组装，避免硬编码
- 变更配置字段时同步更新 `setting/AppConfig` 和 `conf/config.yaml`
- 更新 Swagger 后同步提交 `docs/swagger.yaml` 与 `docs/swagger.json`

