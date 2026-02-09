-- 008_update_context.up.sql
-- Adds new active object IDs to context table for new hierarchy
-- Supports: Ecosystem -> Domain -> App -> Workspace

-- Add new columns for ecosystem, domain, and app context
ALTER TABLE context ADD COLUMN active_ecosystem_id INTEGER REFERENCES ecosystems(id);
ALTER TABLE context ADD COLUMN active_domain_id INTEGER REFERENCES domains(id);
ALTER TABLE context ADD COLUMN active_app_id INTEGER REFERENCES apps(id);

-- Note: active_project_id and active_workspace_id are kept for backward compatibility
-- They will be deprecated in v0.9.0 when Workspace migrates from Project to App
