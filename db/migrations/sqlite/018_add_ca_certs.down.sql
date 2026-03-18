-- Reverse migration: remove ca_certs columns from ecosystems and domains
-- Note: Since SQLite 3.35.0+ supports DROP COLUMN, use that syntax here.

ALTER TABLE ecosystems DROP COLUMN ca_certs;
ALTER TABLE domains DROP COLUMN ca_certs;
