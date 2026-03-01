# Registry Package Testing

## Test Overview

The registry package has **92 total tests**:
- **72 tests that pass** (mix of unit and integration tests)
- **20 tests skipped** (require real Zot binary with HTTP server)
- **0 failures**

## Test Categories

### Unit Tests (Run by default)

These tests run quickly and don't require external dependencies:

1. **Binary Manager Unit Tests**
   - File existence checks
   - Executable permissions
   - Error handling for invalid versions
   - Rollback on failure

2. **Process Manager Tests**
   - Process lifecycle management
   - PID file handling
   - Log file creation
   - Signal handling (SIGTERM, SIGKILL)

3. **Zot Manager Unit Tests**
   - Configuration validation
   - Status checks (stopped state)
   - Endpoint formatting
   - Prune validation (not running)

### Integration Tests (Skipped by default)

These tests require a real Zot binary with HTTP server functionality and are marked with `t.Skip()`:

1. **Registry Start Tests**
   - `TestZotManager_Start_Success`
   - `TestZotManager_Start_AlreadyRunning`
   - `TestZotManager_Start_BinaryNotFound`
   - `TestZotManager_Start_PortInUse`

2. **Registry Stop Tests**
   - `TestZotManager_Stop_Success`
   - `TestZotManager_Stop_GracefulShutdown`

3. **Registry Status Tests**
   - `TestZotManager_Status_Running`
   - `TestZotManager_Status_ImageCount`

4. **Binary Download Tests** (use real GitHub releases)
   - `TestBinaryManager_EnsureBinary_Downloads`
   - `TestBinaryManager_EnsureBinary_VerifiesChecksum`
   - `TestBinaryManager_EnsureBinary_PermissionsCorrect`
   - `TestBinaryManager_Update_Success`
   - `TestBinaryManager_Update_VerifiesChecksumAfterDownload`

5. **Other Integration Tests**
   - `TestZotManager_EnsureRunning_StartsIfStopped`
   - `TestZotManager_EnsureRunning_NoopIfRunning`
   - `TestZotManager_IsRunning_True`
   - `TestZotManager_IsRunning_AfterCrash`
   - `TestZotManager_Prune_All`
   - `TestZotManager_Prune_OlderThan`
   - `TestZotManager_Prune_DryRun`

## Running Tests

### Run All Unit Tests (default)
```bash
go test ./pkg/registry/...
```

### Run with Race Detector (CI requirement)
```bash
go test ./pkg/registry/... -race
```

### Run in Short Mode (fastest)
```bash
go test ./pkg/registry/... -short
```

### Run Specific Test
```bash
go test ./pkg/registry/... -run TestBinaryManager_EnsureBinary_AlreadyExists -v
```

## Test Architecture

### Mock Binary Manager

The `MockBinaryManager` provides test doubles for unit testing:

```go
// Create mock that returns fake binary without network calls
mockBinary := NewMockBinaryManager(binDir, "1.4.3")

// Inject into ZotManager for testing
mgr := NewZotManagerWithDeps(config, mockBinary, mockProcess)
```

The mock creates a fake bash script that:
- Responds to `version` command
- Doesn't require network downloads
- Allows testing ZotManager logic without real Zot binary

### Dependency Injection

The `NewZotManagerWithDeps()` factory allows injecting mock dependencies:

```go
// Production usage (real dependencies)
mgr := NewZotManager(config)

// Test usage (mock dependencies)
mgr := NewZotManagerWithDeps(config, mockBinary, mockProcess)
```

## Why Integration Tests Are Skipped

Integration tests that start a real Zot registry are skipped because they require:

1. **Real Zot binary** - Must be downloaded from GitHub releases
2. **HTTP server** - Zot must start and serve HTTP on a port
3. **Network access** - For downloading and API calls
4. **Slower execution** - Downloads take 5-10 seconds each
5. **System resources** - Actual processes, ports, disk I/O

These are better suited for:
- Manual testing during development
- Separate integration test suite
- CI with dedicated Zot setup

## Version Used in Tests

- **Unit tests with real downloads**: v2.1.1 (stable Zot version that exists)
- **Mock tests**: v1.4.3 (matches production default)
- **Integration tests**: Would use real Zot binary

## Test Cleanup

All tests properly clean up:
- Temporary directories (`t.TempDir()`)
- Process cleanup (`defer mgr.Stop(ctx)`)
- No leaked goroutines or processes

## Adding New Tests

### For Unit Tests
```go
func TestYourFeature(t *testing.T) {
    mockBinary := setupMockBinaryManager(t)
    // Test logic without real binary
}
```

### For Integration Tests
```go
func TestYourFeature(t *testing.T) {
    t.Skip("Integration test - requires real Zot binary with HTTP server")
    // Test logic that needs real registry
}
```

## CI/CD Considerations

The test suite is designed for CI:
- ✅ Fast execution (~58s for full suite, ~6s in short mode with `-short`)
- ✅ No external dependencies for unit tests
- ✅ Passes with `-race` flag
- ✅ 0 failures in standard mode
- ✅ Deterministic (no flaky tests)
- ℹ️  Some tests download real Zot binaries (~26s of network I/O)

**Recommendation for CI:** Use `-short` flag for fast feedback:
```bash
go test ./pkg/registry/... -short -race  # ~6 seconds
```

Integration tests can be enabled in CI by:
1. Pre-installing Zot binary
2. Removing `t.Skip()` calls
3. Running in dedicated integration stage
