CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL, 
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Индексы для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users (user_id) WHERE deleted_at IS NULL;

-- Частичные уникальные индексы (чтобы нельзя было создать двух активных юзеров с одним email/username)
CREATE UNIQUE INDEX users_email_active_key ON users (email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX users_username_active_key ON users (username) WHERE deleted_at IS NULL;