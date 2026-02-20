# Auto-Migration Version Tracking

## Overview

Auto-migration now runs ONLY on first run after an upgrade by tracking the binary version. This provides a fast startup path for users when the version hasn't changed.

## Implementation

### Version Storage
- Last-run binary version is stored in `~/.devopsmaestro/.version`
- Simple text file with version string (e.g., "v0.6.0" or "dev")
- Created automatically on first database operation

### Behavior

#### Fast Path (Version Unchanged)
```bash
# Subsequent runs with same version
./dvm status
# Skips migration check entirely - fast startup
```

#### Migration Path (Version Changed or First Run)
```bash  
# First run after upgrade
./dvm status
# Checks for pending migrations, applies if needed, saves new version
```

### Verbose Logging
```bash
# See what's happening
./dvm --verbose status

# First run: 
# time=... level=INFO msg="first run detected, checking for migrations" version=v0.6.0

# Version change:
# time=... level=INFO msg="version change detected, checking for migrations" old=v0.5.1 new=v0.6.0

# Same version (fast path):
# time=... level=DEBUG msg="version unchanged, skipping migration check" version=v0.6.0
```

### Manual Migration Still Works
```bash
# Force migration regardless of version
./dvm admin migrate
# Uses traditional AutoMigrate function
```

## Files Changed

### `db/database.go`
- Added `CheckVersionBasedAutoMigration()` - main entry point for version-based migration
- Added `GetStoredVersion()` - reads ~/.devopsmaestro/.version
- Added `SaveCurrentVersion()` - writes ~/.devopsmaestro/.version  
- Added `runMigrationsIfNeeded()` - helper for migration logic
- Added comments clarifying AutoMigrate vs CheckVersionBasedAutoMigration usage

### `cmd/root.go`  
- Modified PersistentPreRun to use `CheckVersionBasedAutoMigration()` instead of `AutoMigrate()`
- Passes `Version` variable and `verbose` flag to migration function
- Enhanced error handling and logging

### `db/version_migration_test.go` (new)
- Test coverage for GetStoredVersion() and SaveCurrentVersion()
- Tests first run, version changes, file updates, whitespace handling

## Key Benefits

1. **Fast Startup**: When version unchanged, skips migration check entirely
2. **Automatic Migration**: Still applies migrations when needed after upgrades  
3. **Backward Compatible**: Manual `dvm admin migrate` continues to work
4. **Informative**: Verbose mode shows what's happening
5. **Robust**: Handles edge cases (missing files, permissions, etc.)

## Error Handling

- If version file can't be read: falls back to normal migration check
- If migration fails: version file is NOT updated (so retry happens next run)
- If version save fails: logs warning but doesn't break functionality
- Commands that don't need DB (version, help, completion) skip all migration logic

## Example Scenarios

### New Installation
```bash
./dvm status  # First run
# → Creates ~/.devopsmaestro/.version with current version
# → Runs any needed migrations

./dvm status  # Subsequent runs  
# → Fast path: skips migration check
```

### After Upgrade
```bash
# User upgrades from v0.5.1 to v0.6.0
./dvm status  # First run after upgrade
# → Detects version change (v0.5.1 → v0.6.0) 
# → Runs migration check, applies any new migrations
# → Updates ~/.devopsmaestro/.version to v0.6.0

./dvm status  # Subsequent runs
# → Fast path: skips migration check
```

### Development Version
```bash
# During development with version "dev"
./dvm status  # Every run
# → Version file shows "dev"
# → Still gets fast path behavior during development
```