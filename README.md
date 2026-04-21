# GamePulse

GamePulse 是一个按 `Controller -> Logic -> DAO` 分层组织的 Go + Vue 项目，当前已经实现用户注册登录、社区浏览、发帖、删帖、投票和图片上传等核心社区能力。

在社区主链路之外，当前代码还已经接入了两条异步 AI 链路：

- 发帖后异步调用 LLM 做帖子舆情分析，并把结果落到 PostgreSQL 的 `post_analysis`
- 发帖后异步生成 embedding，并写入 Milvus 向量库

不过，Milvus 侧目前更像“向量检索基础设施已接入”，仓库里还没有对外暴露的相似搜索 API。

## 当前技术栈

- 后端：Go、Gin、PostgreSQL、Redis、MinIO、Milvus
- 前端：Vue 3、Vite、Element Plus
- AI 链路：Qwen 兼容 LLM、DashScope Embedding、Milvus 向量检索

## 主要目录

- `main.go`：后端启动入口
- `routes/`：路由注册
- `controllers/`：接口层
- `logic/`：业务层
- `dao/`：数据访问层
- `dao/milvus/`：Milvus 向量库接入
- `frontend/`：前端源码
- `docs/PROJECT_DOCUMENTATION.md`：当前项目说明文档

## 先看哪里

如果是第一次接手这个仓库，建议先读：

1. `main.go`
2. `routes/routes.go`
3. `docs/PROJECT_DOCUMENTATION.md`

## 说明

仓库中仍然保留了一些旧文档、构建产物和未完全跟上的编排文件。理解项目真实状态时，请优先相信当前代码、`main.go`、`routes/routes.go` 和 `conf/config.yaml`。
