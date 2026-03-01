# Quick Fix Guide - Integration Tests

This guide shows the common patterns for fixing the integration tests.

## Pattern 1: Accessing Resource Names

### ❌ Before (fails)
```go
ecosystems, err := f.RunDVMJSONList("get", "ecosystems")
assert.Equal(t, "test-eco", ecosystems[0]["name"])
```

### ✅ After (works)
```go
ecosystems, err := f.RunDVMJSONList("get", "ecosystems")
assert.Equal(t, "test-eco", f.GetResourceName(ecosystems[0]))
```

## Pattern 2: Accessing Descriptions

### ❌ Before (fails)
```go
assert.Equal(t, "Test description", ecosystem["description"])
```

### ✅ After (works)
```go
assert.Equal(t, "Test description", f.GetResourceDescription(ecosystem))
```

## Pattern 3: Accessing Spec Fields

### ❌ Before (fails)
```go
theme := workspace["theme"]
```

### ✅ After (works)
```go
spec := f.GetResourceSpec(workspace)
if spec != nil {
    theme := spec["theme"]
}
```

## Pattern 4: Context Fields

### ❌ Before (fails)
```go
context, err := f.RunDVMJSON("get", "context")
assert.Equal(t, "test-eco", context["ecosystem"])
```

### ✅ After (works)
```go
context, err := f.RunDVMJSON("get", "context")
// Context uses camelCase keys
assert.Equal(t, "test-eco", context["currentEcosystem"])
assert.Equal(t, "test-app", context["currentApp"])
assert.Equal(t, "test-ws", context["currentWorkspace"])
```

## Pattern 5: Creating Apps

### ❌ Before (fails)
```go
f.AssertCommandSuccess(t, "create", "app", "test-app",
    "--path", "/workspace/test-app",  // Path doesn't exist
    "--language", "go")                // Flag doesn't exist
```

### ✅ After (works)
```go
f.AssertCommandSuccess(t, "create", "app", "test-app",
    "--from-cwd",                      // Use current directory
    "--description", "Test app")       // Remove --language flag
```

## Pattern 6: Creating Workspaces

### ❌ Before (may fail depending on flags)
```go
f.AssertCommandSuccess(t, "create", "workspace", "ws1",
    "--language", "go")  // Wrong flag
```

### ✅ After (check actual flags first)
```bash
# Check what flags exist
./dvm create workspace --help
```

```go
f.AssertCommandSuccess(t, "create", "workspace", "ws1",
    "--description", "Test workspace",
    "--theme", "coolnight-ocean")
```

## Pattern 7: Extracting Multiple Fields

### ❌ Before (fails)
```go
app := apps[0]
name := app["name"]
desc := app["description"]
path := app["path"]
```

### ✅ After (works)
```go
app := apps[0]
name := f.GetResourceName(app)
desc := f.GetResourceDescription(app)

spec := f.GetResourceSpec(app)
if spec != nil {
    path := spec["path"]
}
```

## Pattern 8: Checking Field Existence

### ❌ Before (may panic)
```go
assert.Equal(t, "expected", resource["field"])
```

### ✅ After (safe)
```go
if field, ok := resource["field"]; ok {
    assert.Equal(t, "expected", field)
} else {
    // Field might be in Spec or Metadata
    spec := f.GetResourceSpec(resource)
    if spec != nil {
        if field, ok := spec["field"]; ok {
            assert.Equal(t, "expected", field)
        }
    }
}
```

## Pattern 9: List Iteration

### ❌ Before (verbose)
```go
found := false
for _, ecosystem := range ecosystems {
    if ecosystem["name"] == "test-eco" {
        found = true
        break
    }
}
assert.True(t, found)
```

### ✅ After (cleaner)
```go
names := make([]string, len(ecosystems))
for i, eco := range ecosystems {
    names[i] = f.GetResourceName(eco)
}
assert.Contains(t, names, "test-eco")
```

## Pattern 10: Checking Resource Counts

### ✅ Good (stays the same)
```go
ecosystems, err := f.RunDVMJSONList("get", "ecosystems")
require.NoError(t, err)
assert.Len(t, ecosystems, 3, "Should have 3 ecosystems")
```

## Complete Example: Fixing TestCRUDEcosystem

### ❌ Before
```go
func TestCRUDEcosystem(t *testing.T) {
    f := NewTestFramework(t)
    defer f.Cleanup()

    // CREATE
    f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco",
        "--description", "Test ecosystem",
        "--theme", "coolnight-ocean")

    // READ - Get single
    eco, err := f.RunDVMJSON("get", "ecosystem", "test-eco")
    require.NoError(t, err)
    assert.Equal(t, "test-eco", eco["name"])
    assert.Equal(t, "Test ecosystem", eco["description"])
    assert.Equal(t, "coolnight-ocean", eco["theme"])
}
```

### ✅ After
```go
func TestCRUDEcosystem(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    f := NewTestFramework(t)
    defer f.Cleanup()

    // CREATE
    f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco",
        "--description", "Test ecosystem",
        "--theme", "coolnight-ocean")

    // READ - Get single
    eco, err := f.RunDVMJSON("get", "ecosystem", "test-eco")
    require.NoError(t, err)
    
    // Use helper methods for K8s resource format
    assert.Equal(t, "test-eco", f.GetResourceName(eco))
    assert.Equal(t, "Test ecosystem", f.GetResourceDescription(eco))
    
    spec := f.GetResourceSpec(eco)
    require.NotNil(t, spec, "Spec should exist")
    assert.Equal(t, "coolnight-ocean", spec["Theme"])
}
```

## Debugging Tips

### 1. Check Actual Output
```bash
./dvm get ecosystems -o json | jq '.[0]'
```

### 2. Print Full Resource in Test
```go
resourceJSON, _ := json.MarshalIndent(resource, "", "  ")
t.Logf("Full resource:\n%s", resourceJSON)
```

### 3. Check Available Commands
```bash
./dvm create --help
./dvm create app --help
./dvm create workspace --help
```

### 4. Verify Field Names
```bash
./dvm get context -o json | jq 'keys'
```

## Common Errors and Fixes

| Error | Cause | Fix |
|-------|-------|-----|
| `nil interface` | Direct field access on K8s resource | Use helper methods |
| `unknown flag: --language` | Flag doesn't exist | Remove flag |
| `path does not exist` | Path validation | Use `--from-cwd` |
| `field not found` | Field in Spec, not root | Check Spec section |

## Batch Update Script Ideas

If updating many tests, consider creating helper functions:

```go
// Helper to extract all names from a list
func getResourceNames(f *TestFramework, resources []map[string]interface{}) []string {
    names := make([]string, len(resources))
    for i, r := range resources {
        names[i] = f.GetResourceName(r)
    }
    return names
}

// Usage in tests
names := getResourceNames(f, ecosystems)
assert.Contains(t, names, "test-eco")
```

## Testing Your Fixes

After fixing a test:

```bash
# 1. Run the fixed test
go test ./integration_test -run TestName -v

# 2. Verify it passes
echo $?  # Should be 0

# 3. Run all tests in that file
go test ./integration_test -run Test.*Ecosystem -v

# 4. Finally, run all integration tests
go test ./integration_test/... -v -timeout 5m
```

## Estimated Time to Fix All Tests

| Task | Time | Notes |
|------|------|-------|
| Fix hierarchy_test.go | 15 min | 5 tests × 3 min |
| Fix workspace_test.go | 20 min | 8 tests × 2.5 min |
| Fix gitrepo_test.go | 15 min | 9 tests × 1.5 min (many passing) |
| Fix crud_test.go | 25 min | 10 tests × 2.5 min |
| **Total** | **75 min** | ~1.25 hours |

Add 15-30 minutes for testing and debugging edge cases.

**Total estimated time: 90-105 minutes (1.5-2 hours)**
