CREATE DATABASE Delphine
WITH
    OWNER = admin ENCODING = 'UTF8';

CREATE ROLE delphine_scrapers LOGIN PASSWORD 'securepassword'
CONNECTION
LIMIT 99;

GRANT CONNECT ON DATABASE Delphine TO delphine_scrapers;

CREATE SCHEMA IF NOT EXISTS kb_finance;

GRANT USAGE ON SCHEMA kb_finance TO delphine_scrapers;

GRANT
SELECT,
INSERT,
UPDATE,
DELETE ON ALL TABLES IN SCHEMA kb_finance to delphine_scrapers;