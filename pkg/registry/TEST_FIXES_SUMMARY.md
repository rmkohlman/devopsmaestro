# Registry Test Fixes - Summary

## Problem

The registry tests were failing because they attempted to:
1. Download Zot binaries from non-existent GitHub releases (v1.0.0)
2. Start real Zot processes with HTTP servers
3. Make network calls that took too long or failed

Original failures:
- **20 tests failing** due to HTTP 404 errors and timeouts
- Tests hanging or taking 2+ minutes to complete
- Flaky behavior depending on network conditions

## Solution Applied

### Approach: Mock-Based Unit Tests + Skipped Integration Tests

Combined **Option 1 (Mocks)** and **Option 2 (Skip Integration Tests)** from the task description.

### Changes Made

#### 1. Created MockBinaryManager (`mock_binary_manager.go`)

New file providing a test double for `BinaryManager` interface:
- Creates fake Zot binary (bash script) without network calls
- Responds to `version` command with mock version
- Allows customizable behavior via function hooks
- Fast and deterministic for unit tests

```go
mockBinary := NewMockBinaryManager(binDir, "1.4.3")
mgr := NewZotManagerWithDeps(config, mockBinary, mockProcess)
```

#### 2. Enhanced Factory with Dependency Injection (`factory.go`)

Added `NewZotManagerWithDeps()` for explicit dependency injection:
- Allows injecting mock dependencies in tests
- Maintains backward compatibility with `NewZotManager()`
- Follows SOLID principles (Dependency Inversion)

```go
// Production: uses real dependencies
mgr := NewZotManager(config)

// Testing: injects mocks
mgr := NewZotManagerWithDeps(config, mockBinary, mockProcess)
```

#### 3. Fixed Binary Version in Tests (`binary_manager_test.go`)

- Changed `setupTestBinaryManager()` to use real version **v2.1.1** instead of non-existent v1.0.0
- Added `setupMockBinaryManager()` helper for unit tests
- Tests that download real binaries now use valid version

#### 4. Updated Registry Manager Tests (`registry_manager_test.go`)

- `setupTestRegistryManager()` now injects `MockBinaryManager`
- 14 tests marked with `t.Skip("Integration test - requires real Zot binary with HTTP server")`
- Tests that don't need running registry still pass (status checks, endpoint formatting, etc.)

#### 5. Created Documentation (`TESTING.md`)

Comprehensive test documentation covering:
- Test categories (unit vs integration)
- How to run tests
- Mock architecture
- CI/CD recommendations

## Results

### Test Counts

| Category | Count | Details |
|----------|-------|---------|
| **Total Tests** | 92 | All test functions |
| **Passing** | 72 | Unit + integration tests that run |
| **Skipped** | 20 | Integration tests needing real Zot |
| **Failing** | 0 | ✅ All tests pass |

### Execution Times

| Mode | Time | Details |
|------|------|---------|
| **Full Suite** | ~58s | Includes real binary downloads (v2.1.1) |
| **Short Mode (`-short`)** | ~6s | Skips network tests, fast feedback |
| **With Race Detector** | ~54s | All tests pass with `-race` flag |

### Test Categories

**Unit Tests (run by default):**
- Binary Manager: file operations, permissions, error handling
- Process Manager: lifecycle, signals, PID files
- Zot Manager: configuration, status (stopped), endpoint formatting
- Factory: dependency injection, validation

**Integration Tests (download real binaries):**
- Binary downloads from GitHub (v2.1.1)
- Checksum verification
- Permission checks
- Update with rollback

**Integration Tests (skipped - need HTTP server):**
- Registry start/stop operations
- Status checks on running registry
- Prune operations
- Port conflict detection

## Files Modified

1. **pkg/registry/mock_binary_manager.go** - NEW: Mock implementation
2. **pkg/registry/factory.go** - MODIFIED: Added `NewZotManagerWithDeps()`
3. **pkg/registry/binary_manager_test.go** - MODIFIED: Fixed version to v2.1.1, added mock helper
4. **pkg/registry/registry_manager_test.go** - MODIFIED: Use mocks, skip 14 integration tests
5. **pkg/registry/TESTING.md** - NEW: Comprehensive test documentation

## Verification

```bash
# All tests pass
$ go test ./pkg/registry/...
ok  	devopsmaestro/pkg/registry	58.466s

# Fast mode for CI
$ go test ./pkg/registry/... -short
ok  	devopsmaestro/pkg/registry	6.199s

# Race detector passes
$ go test ./pkg/registry/... -race
ok  	devopsmaestro/pkg/registry	54.521s

# No failures
$ go test ./pkg/registry/... -v 2>&1 | grep "^--- FAIL" | wc -l
0
```

## Benefits

1. **Reliability**: 0 failures, deterministic tests
2. **Speed**: 6s in short mode (10x faster for CI)
3. **Maintainability**: Clear separation of unit vs integration tests
4. **Testability**: Mock injection enables proper unit testing
5. **Documentation**: TESTING.md explains architecture
6. **CI-Friendly**: Fast feedback loop, no external dependencies for unit tests

## Future Improvements

For full integration test coverage:

1. **Add build tags**: Create `*_integration_test.go` files with `//go:build integration`
2. **CI integration stage**: Pre-install Zot binary, enable full tests
3. **Docker-based tests**: Use testcontainers to run real Zot in CI
4. **Mock HTTP server**: Create fake Zot HTTP responses for middleware tests

## Recommendation

For day-to-day development and CI:
```bash
go test ./... -short -race
```

For release verification:
```bash
go test ./... -race
```

For full integration testing:
```bash
# Remove t.Skip() calls or use build tags
go test ./pkg/registry/... -v
```
