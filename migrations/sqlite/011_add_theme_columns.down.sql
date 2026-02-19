-- 011_add_theme_columns.down.sql
-- Removes theme columns from ecosystems, domains, and apps tables

-- Remove theme column from apps table
ALTER TABLE apps DROP COLUMN theme;

-- Remove theme column from domains table
ALTER TABLE domains DROP COLUMN theme;

-- Remove theme column from ecosystems table
ALTER TABLE ecosystems DROP COLUMN theme;