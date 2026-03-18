-- Add ca_certs columns to ecosystems and domains tables
-- for hierarchical CA certificate cascade (v0.56.0)

ALTER TABLE ecosystems ADD COLUMN ca_certs TEXT;
ALTER TABLE domains ADD COLUMN ca_certs TEXT;
