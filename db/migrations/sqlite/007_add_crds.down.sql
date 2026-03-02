-- 007_add_crds.down.sql
-- Remove CRD tables

-- Drop custom_resources indexes
DROP INDEX IF EXISTS idx_cr_kind_name;
DROP INDEX IF EXISTS idx_cr_namespace;
DROP INDEX IF EXISTS idx_cr_name;
DROP INDEX IF EXISTS idx_cr_kind;

-- Drop custom_resources table
DROP TABLE IF EXISTS custom_resources;

-- Drop custom_resource_definitions indexes
DROP INDEX IF EXISTS idx_crd_scope;
DROP INDEX IF EXISTS idx_crd_group;
DROP INDEX IF EXISTS idx_crd_kind;

-- Drop custom_resource_definitions table
DROP TABLE IF EXISTS custom_resource_definitions;
