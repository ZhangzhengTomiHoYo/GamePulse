-- 会话主表
CREATE TABLE conversation (
                                    conversation_id BIGSERIAL PRIMARY KEY,
                                    user_id BIGINT NOT NULL,
                                    agent_code VARCHAR(64) NOT NULL,
                                    title VARCHAR(128) NOT NULL,
                                    message_count INT NOT NULL DEFAULT 0,
                                    last_message_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                    status SMALLINT NOT NULL DEFAULT 1,
                                    create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                    update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_conversation_user ON conversation(user_id, last_message_at DESC);

-- 会话消息表
CREATE TABLE message (
                               message_id BIGSERIAL PRIMARY KEY,
                               conversation_id BIGINT NOT NULL,
                               role VARCHAR(32) NOT NULL, -- 'user', 'assistant'
                               content TEXT NOT NULL,
                               seq BIGINT NOT NULL,       -- 消息序号，用于排序
                               meta_json JSONB,           -- 存 sources, searchQuery 等上下文
                               create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_message_conversation on message(conversation_id, seq ASC);