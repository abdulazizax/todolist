CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    firt_name VARCHAR(255),
    last_name VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(255) DEFAULT 'user',
    refresh_token VARCHAR(255),
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);