-- tables, functions and procedures for the financial knowledge base

--- TABLES ---
CREATE TABLE IF NOT EXISTS kb_finance.providers (                                     
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS kb_finance.articles (                                     
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    modified_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    url TEXT UNIQUE NOT NULL,
    content TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS kb_finance.articles_metadata (
    id UUID PRIMARY KEY                                  
    published_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    author VARCHAR(255) NOT NULL,
    article_url TEXT NOT NULL REFERENCES kb_finance.articles (url),
    provider_id UUID NOT NULL REFERENCES kb_finance.providers (id)
);


-- CREATE INDEX idx_urls_short_url ON urls.urls USING HASH (short_url);


-- ------------
-- -- Get url
-- ------------

-- CREATE OR REPLACE FUNCTION urls.get_url(short_url_p TEXT)
-- RETURNS TEXT AS $$
-- DECLARE
--     url_q TEXT;
-- BEGIN
-- SELECT long_url
-- INTO url_q
-- FROM urls.urls
-- WHERE short_url = short_url_p LIMIT 1;

-- RETURN url_q;
-- END;
-- $$ LANGUAGE plpgsql;


-- ---------------
-- -- Encoding url
-- ---------------

-- CREATE OR REPLACE FUNCTION urls.id_to_short_url(
--     prefix_p TEXT,
--     id_p BIGINT
-- )
-- RETURNS TEXT AS $$
-- DECLARE
--     alphabet TEXT := '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
--     base INT := length(alphabet);
--     encoded_string TEXT := prefix_p;
--     remainder INT;
-- BEGIN
--     IF id_p = 0 THEN
--         RETURN substring(alphabet FROM 1 FOR 1);
-- END IF;

--     WHILE id_p > 0 LOOP
--         remainder := id_p % base;
--         encoded_string := encoded_string || substring(alphabet FROM remainder + 1 FOR 1);
--         id_p := id_p / base;
-- END LOOP;

-- RETURN encoded_string;
-- END;
-- $$ LANGUAGE plpgsql IMMUTABLE;


-- -------------
-- -- Create or replace url
-- -------------

-- CREATE OR REPLACE FUNCTION urls.insert_and_encode_url(
--     prefix_p TEXT,
--     full_url_p TEXT
-- )
-- RETURNS TEXT AS $$
-- DECLARE
--     tmp_short_url TEXT := prefix_p || NOW()::TEXT;
--     new_id BIGINT;
--     short_url_p TEXT;
-- BEGIN
--     -- we insert a new entry in urls, with a temporary value
--     -- and return the id into declared variable
-- INSERT INTO urls.urls (long_url, short_url)
-- VALUES (
--     full_url_p,
--     tmp_short_url) 
-- ON CONFLICT (long_url) DO UPDATE -- requires long_url do have uniue constraint
-- SET short_url = EXCLUDED.short_url
-- RETURNING id INTO new_id;

-- -- we generate the short url
-- short_url_p := urls.id_to_short_url(prefix_p, new_id);

--     -- we update the original url created
-- UPDATE urls.urls
-- SET short_url = short_url_p
-- WHERE id = new_id;

-- RETURN short_url_p;
-- END;
-- $$ LANGUAGE plpgsql;