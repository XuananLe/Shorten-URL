-- Active: 1722584790626@@127.0.0.1@5432@url-shortener

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
                                     user_id UUID PRIMARY KEY 
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
CREATE INDEX idx_shortened on urls(shortened);


-- generate 1000 users
INSERT INTO users (user_id) SELECT uuid_generate_v4() FROM generate_series(1, 1000);

insert into users (user_id) values ('1722584790626@@');

SELECT count(*) from users where user_id = 'adc95590-3972-433c-a3e6-beb425545d12';
-- delete table urls
DELETE FROM urls WHERE original = '234';

-- 1000 times

UPDATE urls SET original = 'https://www.google.com' WHERE original = 'https://www.gooogle.com';

-- seeding 1,000,000 users
INSERT INTO users (user_id) SELECT uuid_generate_v4() FROM generate_series(1, 3000000);

-- bulk insert 1,000,000 users
INSERT INTO users (user_id) SELECT uuid_generate_v4() FROM generate_series(1, 1000000);

-- delete index on users id
DROP INDEX idx_user_id;

INSERT INTO urls (shortened, original, user_id) VALUES (uuid_generate_v4(), 'https://www.google.com', 'adc95590-3972-433c-a3e6-beb425545d12');

-- repeat 1 million times
INSERT INTO urls (shortened, original, user_id) VALUES (uuid_generate_v4(), 'https://www.google.com', 'adc95590-3972-433c-a3e6-beb425545d12');

DO $$
    BEGIN
        FOR i IN 1..1000000 LOOP
                INSERT INTO urls (shortened, original, user_id)
                VALUES (uuid_generate_v4(), 'https://chatgpt.com/c/670ddad5-ec0c-800d-85ec-3b240983e1de', 'adc95590-3972-433c-a3e6-beb425545d12');
            END LOOP;
    END $$;

DO $$
    BEGIN
        FOR i IN 1..1000000 LOOP
                INSERT INTO urls (shortened, original, user_id)
                VALUES (
                           uuid_generate_v4(),
                           'https://chatgpt.com/c/' || substr(md5(random()::text || clock_timestamp()::text || i::text), 1, 10),
                           'adc95590-3972-433c-a3e6-beb425545d12'
                       );
            END LOOP;
    END $$;


SELECT count(*) from urls;
SELECT count(*) from users;

SELECT pg_size_pretty(pg_database_size(current_database())) AS size;
