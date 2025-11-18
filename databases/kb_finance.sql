-- tables, functions and procedures for the financial knowledge base

--- TABLES ---
CREATE TABLE IF NOT EXISTS kb_finance.providers (                                     
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    id UUID PRIMARY KEY,                                
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    author VARCHAR(255) NOT NULL,
    url TEXT UNIQUE NOT NULL,
    provider_id UUID NOT NULL REFERENCES kb_finance.providers (id) 
);

--- FUNCTIONS AND PROCEDURES ---
CREATE OR REPLACE FUNCTION kb_finance.add_provider(name_p VARCHAR(255))
RETURNS UUID AS $$
DECLARE 
    new_id UUID;
BEGIN
    INSERT INTO kb_finance.providers(name)
    VALUES (name_p)
    -- Set on conlict required to trigger the returning clause
    ON CONFLICT (name) DO UPDATE
    SET name = EXCLUDED.name
    RETURNING id INTO new_id;

    RETURN new_id;
END;
$$ LANGUAGE plpgsql;

CREATE TYPE kb_finance.article_metadata_type AS (
    id UUID,
    author VARCHAR,
    provider VARCHAR,
    article_url TEXT,
    published_at TIMESTAMP
);

CREATE OR REPLACE PROCEDURE kb_finance.add_metadata(metadata_p kb_finance.article_metadata_type)
LANGUAGE plpgsql
AS $$
DECLARE provider_id_p UUID;
BEGIN
    provider_id_p := kb_finance.add_provider(metadata_p.provider);

    INSERT INTO kb_finance.articles_metadata(id, published_at, author, url, provider_id)
    VALUES (
        metadata_p.id,
        metadata_p.published_at,
        metadata_p.author,
        metadata_p.article_url,
        provider_id_p
    ) ON CONFLICT (url) DO NOTHING;
END;
$$;

CREATE TYPE kb_finance.article_type AS (
    id UUID,
    url TEXT,
    content TEXT
);

CREATE OR REPLACE PROCEDURE kb_finance.add_article(article_p kb_finance.article_type)
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO kb_finance.articles(id, url, content)
    VALUES (
        article_p.id,
        article_p.url,
        article_p.content
    )
    ON CONFLICT (url) DO UPDATE
    SET content = kb_finance.articles.content || ' ' || EXCLUDED.content;
END;
$$;