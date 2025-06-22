CREATE TABLE "users" (
    serial_id BIGSERIAL PRIMARY KEY,
    id UUID NOT NULL UNIQUE,
    username VARCHAR(32) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_users_deleted_at ON "users"(deleted_at);

CREATE TABLE "thinking_maps" (
    serial_id BIGSERIAL PRIMARY KEY,
    id UUID NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    problem TEXT NOT NULL,
    problem_type VARCHAR(50),
    target TEXT,
    key_points JSONB,
    constraints JSONB,
    conclusion TEXT,
    status INT NOT NULL DEFAULT 1,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_thinking_maps_user_id ON "thinking_maps"(user_id);
CREATE INDEX idx_thinking_maps_deleted_at ON "thinking_maps"(deleted_at);

CREATE TABLE "thinking_nodes" (
    serial_id BIGSERIAL PRIMARY KEY,
    id UUID NOT NULL UNIQUE,
    map_id UUID NOT NULL,
    parent_id UUID,
    node_type VARCHAR(50) NOT NULL,
    question TEXT NOT NULL,
    target TEXT,
    context TEXT DEFAULT '[]',
    conclusion TEXT,
    status INT DEFAULT 0,
    position JSONB DEFAULT '{"x":0,"y":0}',
    metadata JSONB DEFAULT '{}',
    dependencies JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_thinking_nodes_map_id ON "thinking_nodes"(map_id);
CREATE INDEX idx_thinking_nodes_parent_id ON "thinking_nodes"(parent_id);
CREATE INDEX idx_thinking_nodes_deleted_at ON "thinking_nodes"(deleted_at);

CREATE TABLE "node_details" (
    serial_id BIGSERIAL PRIMARY KEY,
    id UUID NOT NULL UNIQUE,
    node_id UUID NOT NULL,
    detail_type VARCHAR(50) NOT NULL,
    content JSONB NOT NULL DEFAULT '{}',
    status INT NOT NULL DEFAULT 1,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_node_details_node_id ON "node_details"(node_id);
CREATE INDEX idx_node_details_deleted_at ON "node_details"(deleted_at);

CREATE TABLE "messages" (
    serial_id BIGSERIAL PRIMARY KEY,
    id UUID NOT NULL UNIQUE,
    node_id UUID NOT NULL,
    parent_id UUID,
    message_type VARCHAR(20) NOT NULL DEFAULT '1',
    content JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_messages_node_id ON "messages"(node_id);
CREATE INDEX idx_messages_parent_id ON "messages"(parent_id);
CREATE INDEX idx_messages_deleted_at ON "messages"(deleted_at);

CREATE TABLE "rag_records" (
    serial_id BIGSERIAL PRIMARY KEY,
    id UUID NOT NULL UNIQUE,
    query TEXT NOT NULL,
    answer TEXT NOT NULL,
    sources JSONB NOT NULL DEFAULT '[]',
    status INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_rag_records_deleted_at ON "rag_records"(deleted_at);