-- 【第一步：注入 AI 灵魂（极度重要）】
CREATE EXTENSION IF NOT EXISTS vector;

-- 【第二步：创建核心表】
-- 1. 用户表
DROP TABLE IF EXISTS "user";
CREATE TABLE "user" (
                        id BIGSERIAL PRIMARY KEY,
                        user_id BIGINT UNIQUE NOT NULL,
                        username VARCHAR(64) UNIQUE NOT NULL,
                        password VARCHAR(64) NOT NULL,
                        email VARCHAR(64),
                        gender SMALLINT NOT NULL DEFAULT 0,
                        create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 社区表
DROP TABLE IF EXISTS community;
CREATE TABLE community (
                           id SERIAL PRIMARY KEY,
                           community_id INT UNIQUE NOT NULL,
                           community_name VARCHAR(128) UNIQUE NOT NULL,
                           introduction VARCHAR(256) NOT NULL,
                           create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. 帖子表
DROP TABLE IF EXISTS post;
CREATE TABLE post (
                      id BIGSERIAL PRIMARY KEY,
                      post_id BIGINT UNIQUE NOT NULL,
                      title VARCHAR(128) NOT NULL,
                      content TEXT NOT NULL,
                      image_url VARCHAR(1024),      -- 新增：MinIO 对象存储链接
                      embedding vector(1536),       -- 新增：AI 文本向量化存储
                      author_id BIGINT NOT NULL,
                      community_id BIGINT NOT NULL,
                      status SMALLINT NOT NULL DEFAULT 1,
                      create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                      update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 【第三步：建立索引提升检索速度】
CREATE INDEX idx_post_author_id ON post(author_id);
CREATE INDEX idx_post_community_id ON post(community_id);

-- 【第四步：灌入初始社区数据】
INSERT INTO community (community_id, community_name, introduction, create_time, update_time) VALUES
                                                                                                 (1, '原神', '提瓦特大陆的奇幻冒险太治愈！丰富剧情、精美角色与开放世界探索，是无数玩家心中的“快乐老家”', '2026-01-06 10:00:00', '2026-01-06 10:00:00'),
                                                                                                 (2, '崩坏：星穹铁道', '星穹列车的银河冒险超精彩！科幻剧情+策略回合制战斗，每个星神阵营的设定都让人疯狂心动', '2026-01-06 11:30:00', '2026-01-06 11:30:00'),
                                                                                                 (3, '绝区零', '新艾利都的“绳匠”委托太有趣！快节奏战斗+赛博都市氛围，角色的反差萌直接戳中喜好', '2026-01-06 13:15:00', '2026-01-06 13:15:00'),
                                                                                                 (4, '鸣潮', '启明星群岛的共鸣冒险超震撼！开放世界+流畅战斗手感，每个区域的生态设计都让人想反复探索', '2026-01-06 14:40:00', '2026-01-06 14:40:00'),
                                                                                                 (5, '明日方舟', '罗德岛的干员们太靠谱！塔防+剧情深度拉满，每个干员的背景故事都让人又爱又心疼', '2026-01-06 16:20:00', '2026-01-06 16:20:00'),
                                                                                                 (6, '碧蓝航线', '舰娘们的日常与海战超治愈！立绘精美+养成轻松，每次活动的剧情都甜到心坎里', '2026-01-06 17:50:00', '2026-01-06 17:50:00'),
                                                                                                 (7, '崩坏3', '女武神们的终焉叙事太动人！动作战斗的打击感+情怀剧情，是陪伴无数玩家成长的青春回忆', '2026-01-06 19:10:00', '2026-01-06 19:10:00');