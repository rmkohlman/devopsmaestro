-- 003_add_git_repo_fk.up.sql
-- Add git_repo_id foreign key to apps and workspaces tables

ALTER TABLE apps ADD COLUMN git_repo_id INTEGER REFERENCES git_repos(id) ON DELETE SET NULL;
ALTER TABLE workspaces ADD COLUMN git_repo_id INTEGER REFERENCES git_repos(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_apps_git_repo ON apps(git_repo_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_git_repo ON workspaces(git_repo_id);
