-- Active: 1722584790626@@127.0.0.1@5432@url-shortener@public

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY  -- Removed the trailing comma
);

CREATE TABLE IF NOT EXISTS urls (  
    shortened VARCHAR(100) PRIMARY KEY,                
    original VARCHAR(250) NOT NULL,
    clicks BIGINT DEFAULT 0,            
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, 
    expired_at TIMESTAMPTZ DEFAULT (CURRENT_TIMESTAMP + INTERVAL '100 days'), -- New default value
    user_id UUID,                                   
    CONSTRAINT fk_url_user_id FOREIGN KEY (user_id) REFERENCES users(user_id)  
);

-- Indexes
CREATE INDEX idx_original ON urls(original);  
CREATE INDEX idx_user_id ON urls(user_id);

CREATE INDEX idx_shortened ON urls(shortened);
