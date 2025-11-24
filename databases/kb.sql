-- tables, functions and procedures for the financial knowledge base

--- TABLES ---
CREATE TABLE IF NOT EXISTS kb.articles (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    modified_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    url TEXT UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    published TIMESTAMP NOT NULL,
    provider VARCHAR(255) NOT NULL,
    domain VARCHAR(100) NOT NULL,
    content TEXT NOT NULL
);

--- FUNCTIONS AND PROCEDURES ---
CREATE TYPE kb.article_type AS (
    id UUID,
    url TEXT,
    title VARCHAR,
    author VARCHAR,
    published TIMESTAMP,
    provider VARCHAR,
    domain VARCHAR,
    content TEXT
);

CREATE OR REPLACE PROCEDURE kb.add_article(article_p kb.article_type)
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO kb.articles(id, url, title, author, published, provider, domain, content)
    VALUES (
        article_p.id,
        article_p.url,
        article_p.title,
        article_p.author,
        article_p.published,
        article_p.provider,
        article_p.domain,
        article_p.content
    )
    ON CONFLICT (url) DO NOTHING;
END;
$$;