-- Add build args columns to workspaces, ecosystems, and domains tables
-- for hierarchical build args cascade (v0.55.0)

ALTER TABLE workspaces ADD COLUMN build_config TEXT;
ALTER TABLE ecosystems ADD COLUMN build_args TEXT;
ALTER TABLE domains ADD COLUMN build_args TEXT;
