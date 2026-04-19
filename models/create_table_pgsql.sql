-- ============================================================================
-- 游戏社区舆情监控系统 - 核心数据库初始化脚本 (PostgreSQL + PGVector)
-- 包含：业务真源层 (OLTP) -> 分析特征层 (JSONB) -> 向量检索层 (Vector)
-- ============================================================================

-- 【前置准备：注入 AI 灵魂】
CREATE EXTENSION IF NOT EXISTS vector;

-- 【清理旧表（注意顺序，先删依赖表）】
DROP TABLE IF EXISTS post_embeddings CASCADE;
DROP TABLE IF EXISTS post_analysis CASCADE;
DROP TABLE IF EXISTS post CASCADE;
DROP TABLE IF EXISTS community CASCADE;
DROP TABLE IF EXISTS "user" CASCADE;

-- ==========================================
-- 第一层：业务真源表 (OLTP，要求极速响应)
-- ==========================================

-- 1. 用户表
CREATE TABLE "user" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(64) UNIQUE NOT NULL,
    password VARCHAR(64) NOT NULL,
    email VARCHAR(64),
    gender SMALLINT NOT NULL DEFAULT 0,
    create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. 社区表
CREATE TABLE community (
    id SERIAL PRIMARY KEY,
    community_id INT UNIQUE NOT NULL,
    community_name VARCHAR(128) UNIQUE NOT NULL,
    introduction VARCHAR(256) NOT NULL,
    create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. 帖子主表 (只存核心业务，不存向量，保障发帖接口 TPS)
CREATE TABLE post (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT UNIQUE NOT NULL,  -- 雪花算法生成的全局唯一ID
    title VARCHAR(128) NOT NULL,
    content TEXT NOT NULL,           -- 使用 TEXT 防止超长内容报错
    image_url VARCHAR(1024),         -- MinIO 对象存储链接
    author_id BIGINT NOT NULL,
    community_id BIGINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1, -- 1:正常, 0:已删除/隐藏
    create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 主表业务索引
CREATE INDEX idx_post_author_id ON post(author_id);
-- 联合索引：社区帖子列表通常按时间倒序排列
CREATE INDEX idx_post_community_time ON post(community_id, create_time DESC);


-- ==========================================
-- 第二层：分析特征表 (用于 BI 报表、风险告警)
-- ==========================================

-- 4. 帖子分析结果表 (由大模型生成 JSON 后写入)
CREATE TABLE post_analysis (
    post_id BIGINT PRIMARY KEY,      -- 1:1 绑定帖子主表

    -- 情绪倾向
    sentiment_label VARCHAR(16) NOT NULL,   -- 枚举: 'positive', 'neutral', 'negative'
    sentiment_score DECIMAL(3,2),           -- 情感得分: -1.00 到 1.00

    -- 风险控制
    risk_level SMALLINT NOT NULL DEFAULT 0, -- 0:无风险, 1:吐槽, 2:引战, 3:公关危机, 4:违规

    -- 结构化特征 (JSONB 类型，极度方便后续做统计和搜索)
    topics JSONB,     -- 提取的话题，如: ["角色强度", "抽卡概率"]
    keywords JSONB,   -- 提取的关键词，如: ["削弱", "骗氪"]
    summary TEXT,     -- 内容摘要 (50字以内)

    analyzed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束：主表删帖时，级联删除分析数据
    CONSTRAINT fk_analysis_post FOREIGN KEY (post_id) REFERENCES post(post_id) ON DELETE CASCADE
);

-- 分析表索引
CREATE INDEX idx_analysis_sentiment ON post_analysis (sentiment_label);
CREATE INDEX idx_analysis_risk ON post_analysis (risk_level);
-- GIN 索引：支持极速查找包含某个特定 topic 的所有帖子
CREATE INDEX idx_analysis_topics ON post_analysis USING GIN (topics);


-- ==========================================
-- 第三层：向量检索表 (用于相似召回、语义去重)
-- ==========================================

-- 5. 帖子文本切片向量表 (长文本分块，1:N 映射)
CREATE TABLE post_embeddings (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,

    -- 分块控制 (极其重要：解决长文本无法一次性 Embedding 的限制)
    chunk_index INT NOT NULL DEFAULT 0,
    chunk_text TEXT,                     -- 建议保留切片文本，召回时直接展示

    -- 混合检索的"防腐/冗余字段" (用于在算距离前，极速圈定范围)
    community_id BIGINT NOT NULL,
    post_create_time TIMESTAMP WITH TIME ZONE NOT NULL,

    -- 模型与指纹
    model_name VARCHAR(64) NOT NULL,
    model_version VARCHAR(32) NOT NULL,
    content_hash VARCHAR(64) NOT NULL,

    -- 向量本体 (OpenAI 使用 1536，BGE-large 使用 1024。这里以 OpenAI 为例)
    embedding VECTOR(1536),

    -- 异步队列状态
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    error_msg TEXT,

    create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 级联清理：帖子删除时，自动清理底层的所有向量分块
    CONSTRAINT fk_post_embeddings_post
        FOREIGN KEY (post_id) REFERENCES post(post_id) ON DELETE CASCADE,

    -- 唯一键：同一个帖子、同一个切片、同一套模型只能算一次
    CONSTRAINT uq_post_embedding_unique
        UNIQUE (post_id, chunk_index, model_name, model_version)
);

-- 向量表混合检索索引
-- 1. B-Tree：用于 WHERE 条件的极速过滤（选定游戏、时间范围）
CREATE INDEX idx_post_embeddings_community_time ON post_embeddings (community_id, post_create_time DESC);
-- 2. HNSW：用于高维向量空间的余弦距离 (<=>) 近似计算
CREATE INDEX idx_post_embeddings_hnsw ON post_embeddings USING hnsw (embedding vector_cosine_ops);


-- ==========================================
-- 第四步：灌入初始基础数据
-- ==========================================

INSERT INTO community (community_id, community_name, introduction, create_time, update_time) VALUES
    (1, '原神', '提瓦特大陆的奇幻冒险太治愈！丰富剧情、精美角色与开放世界探索，是无数玩家心中的“快乐老家”', '2026-01-06 10:00:00', '2026-01-06 10:00:00'),
    (2, '崩坏：星穹铁道', '星穹列车的银河冒险超精彩！科幻剧情+策略回合制战斗，每个星神阵营的设定都让人疯狂心动', '2026-01-06 11:30:00', '2026-01-06 11:30:00'),
    (3, '绝区零', '新艾利都的“绳匠”委托太有趣！快节奏战斗+赛博都市氛围，角色的反差萌直接戳中喜好', '2026-01-06 13:15:00', '2026-01-06 13:15:00'),
    (4, '鸣潮', '索拉里斯大陆的共鸣冒险超震撼！开放世界+流畅战斗手感，快来击败鸣式，拯救世界', '2026-01-06 14:40:00', '2026-01-06 14:40:00'),
    (5, '明日方舟', '罗德岛的干员们太靠谱！塔防+剧情深度拉满，每个干员的背景故事都让人又爱又心疼', '2026-01-06 16:20:00', '2026-01-06 16:20:00'),
    (6, '碧蓝航线', '舰娘们的日常与海战超治愈！立绘精美+养成轻松，每次活动的剧情都甜到心坎里', '2026-01-06 17:50:00', '2026-01-06 17:50:00'),
    (7, '崩坏3', '女武神们的终焉叙事太动人！动作战斗的打击感+情怀剧情，是陪伴无数玩家成长的青春回忆', '2026-01-06 19:10:00', '2026-01-06 19:10:00');
