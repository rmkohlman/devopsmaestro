-- 011_add_theme_columns.up.sql
-- Adds theme column to ecosystems, domains, and apps tables for hierarchical theme inheritance
-- Theme values are theme names (e.g., "coolnight-ocean") or NULL for inheritance

-- Add theme column to ecosystems table
ALTER TABLE ecosystems ADD COLUMN theme TEXT;

-- Add theme column to domains table  
ALTER TABLE domains ADD COLUMN theme TEXT;

-- Add theme column to apps table
ALTER TABLE apps ADD COLUMN theme TEXT;