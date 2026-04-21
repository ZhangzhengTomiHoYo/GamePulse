# GamePulse 项目文档

更新时间：2026-04-21  
适用代码基线：当前仓库源码  
项目路径：`E:\BaiduSyncdisk\Golang\GamePulse\Project\GamePulse`

## 1. 项目定位

`GamePulse` 的 Go 模块名是 `bluebell`。这是一个按 CLD 思路组织的社区项目：

- `Controller` 负责 HTTP 参数绑定、鉴权上下文读取和统一响应
- `Logic` 负责业务编排
- `DAO` 负责 PostgreSQL、Redis、MinIO 等底层数据访问

当前代码已经实现的是“游戏社区 / 游戏博客”的主体能力，并且已经接入了异步舆情分析与向量化能力：

- 发帖成功后，异步调用 LLM 生成帖子分析结果
- 发帖成功后，异步生成 embedding 并写入 Milvus
- 帖子详情和列表接口会附带 `sentiment_label`

但“舆情监控”仍未完全发展成独立产品能力，目前更准确的状态是：

- 情感分析已经落地到发帖后的异步处理链路
- Milvus 向量检索基础设施已经接入
- 对外的相似搜索 / 语义召回接口还没有暴露
- 定时监控、告警、看板类能力还没有形成闭环

## 2. 当前实现状态

### 已实现的后端能力

- 用户注册：`POST /api/v1/signup`
- 用户登录：`POST /api/v1/login`
- 社区列表：`GET /api/v1/community`
- 社区详情：`GET /api/v1/community/:id`
- 发帖：`POST /api/v1/post`
- 删帖：`DELETE /api/v1/post/:id`
- 帖子详情：`GET /api/v1/post/:id`
- 帖子列表（旧版）：`GET /api/v1/posts`
- 帖子列表（新版，支持按时间或热度排序、按社区筛选）：`GET /api/v1/posts2`
- 帖子投票：`POST /api/v1/vote`
- 图片上传：`POST /api/v1/upload`
- 诊断与文档：`GET /ping`、`GET /swagger/*any`、`pprof`

除注册和登录外，其余业务接口都挂在 JWT 中间件之后。

### 已实现的 AI / 分析链路

- 发帖后异步舆情分析：`controllers.CreatePostHandler -> logic.AnalyzePostAsync -> logic.analyzeAndSavePost -> pgsql.UpsertPostAnalysis`
- 发帖后异步向量化：`controllers.CreatePostHandler -> logic.EmbedPostAsync -> logic.embedAndSavePost -> milvus.UpsertSinglePostVector`
- 帖子详情和列表附带情绪标签：`logic.attachPostAnalysis`
- 服务启动时自动初始化 Milvus，并在需要时创建数据库、集合和索引：`main.go -> dao/milvus.Init -> dao/milvus.EnsureCollection`

### 已接好的前端页面

- 登录 / 注册页：`frontend/src/views/LoginView.vue`
- 首页 feed：`frontend/src/views/HomeView.vue`
- 发帖页：`frontend/src/views/CreatePostView.vue`
- 帖子详情页：`frontend/src/views/PostDetailView.vue`

### 尚未形成完整闭环的目标能力

- 评论系统
- 对外可用的 Milvus 相似搜索接口
- 定时化舆情监控任务流
- 告警、聚合报表和看板

也就是说，“分析”已经不是纯规划，但“监控平台化”还没有完成。

## 3. 技术栈

### 后端

- Go `1.25`
- `gin-gonic/gin`
- `jmoiron/sqlx`
- PostgreSQL 驱动：`lib/pq`
- Redis：`go-redis`
- MinIO：`minio-go/v7`
- Milvus：`milvus-io/milvus/client/v2`
- JWT：`dgrijalva/jwt-go`
- 配置：`viper`
- 日志：`zap` + `lumberjack`
- 文档：`swaggo/gin-swagger`
- 调试：`gin-contrib/pprof`
- 限流：`juju/ratelimit`

### AI / 分析链路

- LLM 调用：`cloudwego/eino-ext/components/model/openai`
- Embedding：`cloudwego/eino-ext/components/embedding/dashscope`
- 向量检索：Milvus
- 结构化分析落库：PostgreSQL `post_analysis`

### 前端

- Vue 3
- Vite
- Element Plus
- axios
- vue-router

## 4. 代码结构

### 后端主干

- `main.go`
  负责初始化配置、日志、PostgreSQL、Redis、MinIO、Snowflake、校验器翻译器，并启动 HTTP 服务。

- `routes/routes.go`
  注册所有路由、中间件、Swagger 和 pprof。

- `controllers/`
  接收请求、校验参数、调用 logic、输出统一响应。

- `logic/`
  聚合数据库、缓存和对象存储操作，封装业务流程。

- `dao/pgsql/`
  PostgreSQL 数据访问层。

- `dao/redis/`
  排序、社区帖子集合、投票记录、缓存清理。

- `dao/minio/`
  MinIO 初始化和文件上传。

- `dao/milvus/`
  Milvus 初始化、集合校验、索引创建、向量写入与相似检索基础能力。

- `models/`
  请求参数结构、数据库模型、接口响应模型、建表 SQL。

- `middlewares/`
  JWT 鉴权和全局限流。

- `pkg/`
  通用能力，目前主要是 JWT 和 Snowflake。

- `logic/llm.go`
  发帖后的异步舆情分析流程。

- `logic/embed.go`
  发帖后的异步 embedding 生成与 Milvus upsert 流程。

### 前端主干

- `frontend/src/main.js`
  Vue 应用入口。

- `frontend/src/router/index.js`
  前端页面路由。

- `frontend/src/api/post.js`
  发帖、删帖、上传图片相关接口封装。

- `frontend/src/views/`
  页面组件源码。

### 构建产物与非源码目录

- `frontend/dist/`：前端构建输出
- `assets/`、`templates/index.html`：后端直接提供的前端静态资源
- `docs/swagger.json`、`docs/swagger.yaml`：Swagger 生成产物
- `tmp/`、`web_app.log`：运行产物

阅读和修改业务代码时，优先看 `frontend/src/`，不要把根目录下的静态构建产物当成源码。

## 5. 请求链路与分层职责

请求的主链路是：

`Router -> Middleware -> Controller -> Logic -> DAO -> Storage`

职责分工如下：

- `Controller`
  负责绑定请求参数、读取当前用户 ID、返回统一 JSON 结构。

- `Logic`
  负责把多个 DAO 调用串起来，完成业务编排。

- `DAO`
  负责访问具体的数据源，不承载上层业务流程。

一个典型例子是发帖：

1. `routes/routes.go` 注册 `/api/v1/post`
2. `controllers.CreatePostHandler` 绑定参数并组装 `models.Post`
3. `logic.CreatePost` 生成帖子 ID
4. `dao/pgsql.CreatePost` 写入 PostgreSQL
5. `dao/redis.CreatePost` 刷新排序和社区集合

## 6. 关键业务说明

### 认证

- 登录成功后返回 `token`
- 业务接口要求请求头携带：`Authorization: Bearer <token>`
- JWT 中间件位于 `middlewares/auth.go`
- token 的生成和解析位于 `pkg/jwt/jwt.go`

### 帖子

- 帖子主数据存储在 PostgreSQL
- 帖子图片 URL 以 JSON 字符串形式存储在 `post.image_url`
- 返回给前端时，会在 logic 层反序列化成 `image_urls`
- 删除帖子采用软删除，且会同步清理 Redis 相关缓存
- 发帖成功后不会阻塞等待 AI 分析结果，而是异步触发分析和 embedding

### 排序与投票

- Redis 使用有序集合维护按时间和按热度排序的帖子列表
- 社区筛选通过社区集合与排序集合做交集
- 投票数据也保存在 Redis 中
- 帖子详情和列表会聚合 PostgreSQL 的帖子数据与 Redis 的投票数据

### 图片上传

- 上传入口是 `POST /api/v1/upload`
- 逻辑层会校验扩展名白名单
- 实际文件保存到 MinIO
- 返回前端的是可公开访问的 URL

### 异步舆情分析

- 触发点：`controllers.CreatePostHandler`
- 模型调用封装在 `logic/llm.go`
- 结果落到 PostgreSQL 的 `post_analysis`
- 列表和详情接口会把 `sentiment_label` 挂回 `ApiPostDetail`

### 向量化与 Milvus

- 触发点：`controllers.CreatePostHandler`
- Embedding 生成封装在 `logic/embed.go`
- 当前配置使用 DashScope embedding，维度是 `1024`
- 向量写入 Milvus 集合，当前默认集合名是 `post_vectors`
- Milvus 侧已经有 `SearchSimilar` 基础能力，但当前没有 route/controller 暴露该能力

## 7. 数据存储

### PostgreSQL

当前实际代码使用的是 PostgreSQL，不是 MySQL。

关键表：

- `"user"`：用户
- `community`：社区
- `post`：帖子
- `post_analysis`：帖子分析结果
- `post_embeddings`：帖子 embedding 分块表

表结构参考：

- `models/create_table_pgsql.sql`

注意点：

- 用户密码当前仍使用 `md5(secret + password)` 方式处理，仅适合学习和演示，不适合生产
- `post_analysis` 当前已经在实际业务链路中被写入和读取
- `post_embeddings` 的表结构和 DAO 已存在，但当前主流程里的 embedding 写入目标是 Milvus，不是这张表
- `models/create_table_pgsql.sql` 里 `post_embeddings.embedding` 使用 `VECTOR(1536)`，而当前 `embedding` 和 `milvus` 配置维度是 `1024`，两者并不一致

### Redis

Redis 主要承担：

- 帖子时间排序
- 帖子热度排序
- 社区帖子集合
- 单帖投票记录

### MinIO

MinIO 负责帖子图片上传，配置位于 `conf/config.yaml` 的 `minio` 段。

### Milvus

Milvus 是当前实际接入的向量数据库，负责帖子向量写入和相似检索基础能力。

当前代码中的 Milvus 特点：

- 启动时尝试初始化
- 自动校验数据库与集合是否存在
- 自动创建向量索引和标量索引
- 发帖后异步 upsert 向量
- 暂未对外暴露检索接口

## 8. 配置说明

配置文件：

- `conf/config.yaml`

当前代码实际读取的配置包括：

- `name`
- `mode`
- `port`
- `version`
- `start_time`
- `machine_id`
- `log`
- `postgres`
- `redis`
- `minio`
- `llm`
- `milvus`
- `embedding`

其中数据源已经是：

- PostgreSQL
- Redis
- MinIO
- Milvus

## 9. 本地运行建议

### 依赖

后端当前依赖以下服务：

- PostgreSQL
- Redis
- MinIO
- Milvus

### 初始化

1. 准备 PostgreSQL、Redis、MinIO、Milvus
2. 在 PostgreSQL 中执行 `models/create_table_pgsql.sql`
3. 根据本机环境修改 `conf/config.yaml`
4. 为 `llm` 和 `embedding` 配置可用的 API Key
5. 启动后端：`go run .`

### 前端开发

1. 进入 `frontend/`
2. 安装依赖：`npm install`
3. 启动开发服务：`npm run dev`

开发模式下，`frontend/vite.config.js` 会把 `/api` 请求代理到 `http://localhost:8080`。

补充说明：

- 如果 Milvus 初始化失败，服务目前会记录 `warn` 并继续启动
- 这意味着基础社区功能仍可运行，但 embedding / 语义检索相关能力会降级

## 10. 当前文档与代码的已知偏差

以下内容在仓库中仍然存在，但与当前代码不完全一致：

- 旧版文档里仍有 MySQL 相关描述
- `docker-compose.yml` 仍保留 `mysql` 服务定义，尚未跟上 PostgreSQL、MinIO、Milvus 的当前方案
- 根目录 `README.md` 原本几乎为空

因此在判断系统真实状态时，优先级建议为：

1. `go.mod`
2. `main.go`
3. `routes/routes.go`
4. `setting/setting.go`
5. 对应的 controller / logic / dao 文件
6. `conf/config.yaml`
7. 再参考本文档和 Swagger

## 11. 已知问题与风险

- `controllers/vote.go` 在参数校验失败后的一个分支缺少 `return`
- 密码算法仍使用 MD5 + 固定盐
- JWT 密钥写死在代码里
- `docker-compose.yml` 仍未和当前 PostgreSQL / MinIO / Milvus 方案完全对齐
- 仓库里保留了较多构建产物和日志产物，容易干扰阅读
- 终端中部分中文注释和字符串可能出现乱码显示，这更像是控制台编码问题，不一定是源码本身损坏
- Milvus 已接入但暂未暴露 controller / route 级的相似检索接口
- PostgreSQL `post_embeddings` 表结构维度是 `1536`，而当前 embedding / Milvus 配置维度是 `1024`
- `dao/pgsql/post_embedding.go` 已存在，但当前活跃的 embedding 主链路并没有写这张表

## 12. 阅读建议

### 想快速理解项目时

按下面顺序读：

1. `main.go`
2. `setting/setting.go`
3. `conf/config.yaml`
4. `routes/routes.go`
5. 再进入具体功能对应的 `controllers/`、`logic/`、`dao/`

### 想改某个功能时

优先沿着这条线找：

`路由 -> controller -> logic -> dao -> models -> 前端页面/接口`

### 想判断某个功能是否真的实现时

同时确认下面三件事：

1. 有没有路由或前端入口
2. 有没有 logic 层编排
3. 有没有真实数据源写入或读取

只有整条链路都存在，才算真正实现。
