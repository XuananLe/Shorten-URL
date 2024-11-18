-- Create schema public
CREATE SCHEMA IF NOT EXISTS public;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the users table
CREATE TABLE IF NOT EXISTS users (
                                     user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
) PARTITION BY HASH (user_id);

-- Create the partitioned urls table
CREATE TABLE IF NOT EXISTS urls (
                                    shortened VARCHAR(100),
                                    original VARCHAR(250) NOT NULL,
                                    clicks BIGINT DEFAULT 0,
                                    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
                                    expired_at TIMESTAMPTZ DEFAULT (CURRENT_TIMESTAMP + INTERVAL '100 days'),
                                    user_id UUID,
                                    CONSTRAINT pk_urls PRIMARY KEY (shortened, user_id),  -- Composite primary key
                                    CONSTRAINT fk_url_user_id FOREIGN KEY (user_id) REFERENCES users(user_id)
) PARTITION BY HASH (shortened);



CREATE TABLE urls_p0 PARTITION OF urls FOR VALUES WITH (MODULUS 5, REMAINDER 0);
CREATE TABLE urls_p1 PARTITION OF urls FOR VALUES WITH (MODULUS 5, REMAINDER 1);
CREATE TABLE urls_p2 PARTITION OF urls FOR VALUES WITH (MODULUS 5, REMAINDER 2);
CREATE TABLE urls_p3 PARTITION OF urls FOR VALUES WITH (MODULUS 5, REMAINDER 3);
CREATE TABLE urls_p4 PARTITION OF urls FOR VALUES WITH (MODULUS 5, REMAINDER 4);


-- Create the index for the urls table
CREATE INDEX idx_urls_0 ON urls_p0 (shortened);
CREATE INDEX idx_urls_1 ON urls_p1 (shortened);
CREATE INDEX idx_urls_2 ON urls_p2 (shortened);
CREATE INDEX idx_urls_3 ON urls_p3 (shortened);
CREATE INDEX idx_urls_4 ON urls_p4 (shortened);


CREATE TABLE users_0 PARTITION OF users FOR VALUES WITH (MODULUS 5, REMAINDER 0);
CREATE TABLE users_1 PARTITION OF users FOR VALUES WITH (MODULUS 5, REMAINDER 1);
CREATE TABLE users_2 PARTITION OF users FOR VALUES WITH (MODULUS 5, REMAINDER 2);
CREATE TABLE users_3 PARTITION OF users FOR VALUES WITH (MODULUS 5, REMAINDER 3);
CREATE TABLE users_4 PARTITION OF users FOR VALUES WITH (MODULUS 5, REMAINDER 4);

CREATE INDEX idx_users_0 ON users_0 (user_id);
CREATE INDEX idx_users_1 ON users_1 (user_id);
CREATE INDEX idx_users_2 ON users_2 (user_id);
CREATE INDEX idx_users_3 ON users_3 (user_id);
CREATE INDEX idx_users_4 ON users_4 (user_id);




