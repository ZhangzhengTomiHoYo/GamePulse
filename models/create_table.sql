CREATE TABLE `user` (
                        `id` bigint(20) NOT NULL AUTO_INCREMENT,
                        `user_id` bigint(20) NOT NULL,
                        `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
                        `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
                        `email` varchar(64) COLLATE utf8mb4_general_ci,
                        `gender` tinyint(4) NOT NULL DEFAULT '0',
                        `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                        `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `idx_username` (`username`) USING BTREE,
                        UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

DROP TABLE IF EXISTS `community`;
CREATE TABLE `community` (
                             `id` int(11) NOT NULL AUTO_INCREMENT,
                             `community_id` int(10) unsigned NOT NULL,
                             `community_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,
                             `introduction` varchar(256) COLLATE utf8mb4_general_ci NOT NULL,
                             `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `idx_community_id` (`community_id`),
                             UNIQUE KEY `idx_community_name` (`community_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `community` VALUES ('1', '1', '原神', '提瓦特大陆的奇幻冒险太治愈！丰富剧情、精美角色与开放世界探索，是无数玩家心中的“快乐老家”', '2026-01-06 10:00:00', '2026-01-06 10:00:00');
INSERT INTO `community` VALUES ('2', '2', '崩坏：星穹铁道', '星穹列车的银河冒险超精彩！科幻剧情+策略回合制战斗，每个星神阵营的设定都让人疯狂心动', '2026-01-06 11:30:00', '2026-01-06 11:30:00');
INSERT INTO `community` VALUES ('3', '3', '绝区零', '新艾利都的“绳匠”委托太有趣！快节奏战斗+赛博都市氛围，角色的反差萌直接戳中喜好', '2026-01-06 13:15:00', '2026-01-06 13:15:00');
INSERT INTO `community` VALUES ('4', '4', '鸣潮', '启明星群岛的共鸣冒险超震撼！开放世界+流畅战斗手感，每个区域的生态设计都让人想反复探索', '2026-01-06 14:40:00', '2026-01-06 14:40:00');
INSERT INTO `community` VALUES ('5', '5', '明日方舟', '罗德岛的干员们太靠谱！塔防+剧情深度拉满，每个干员的背景故事都让人又爱又心疼', '2026-01-06 16:20:00', '2026-01-06 16:20:00');
INSERT INTO `community` VALUES ('6', '6', '碧蓝航线', '舰娘们的日常与海战超治愈！立绘精美+养成轻松，每次活动的剧情都甜到心坎里', '2026-01-06 17:50:00', '2026-01-06 17:50:00');
INSERT INTO `community` VALUES ('7', '7', '崩坏3', '女武神们的终焉叙事太动人！动作战斗的打击感+情怀剧情，是陪伴无数玩家成长的青春回忆', '2026-01-06 19:10:00', '2026-01-06 19:10:00');

DROP TABLE IF EXISTS `post`;
CREATE TABLE `post` (
                        `id` bigint(20) NOT NULL AUTO_INCREMENT,
                        `post_id` bigint(20) NOT NULL COMMENT '帖子id',
                        `title` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
                        `content` varchar(8192) COLLATE utf8mb4_general_ci NOT NULL COMMENT '内容',
                        `author_id` bigint(20) NOT NULL COMMENT '作者的用户id',
                        `community_id` bigint(20) NOT NULL COMMENT '所属社区',
                        `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '帖子状态',
                        `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                        `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `idx_post_id` (`post_id`),
                        KEY `idx_author_id` (`author_id`),
                        KEY `idx_community_id` (`community_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;