# Build Architecture Documentation

## DevOpsMaestro Build System

### Two-Layer Architecture

DevOpsMaestro uses a **hybrid approach** for container operations:

```
┌─────────────────────────────────────────────────┐
│           Build Operations (CLI)                │
│                                                 │
│  ┌─────────────────────────────────────────┐  │
│  │   colima nerdctl build                  │  │
│  │   - Simple subprocess calls             │  │
│  │   - Full BuildKit features              │  │
│  │   - Progress output                     │  │
│  │   - Multi-stage builds                  │  │
│  └─────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│        Runtime Operations (API)                 │
│                                                 │
│  ┌─────────────────────────────────────────┐  │
│  │   containerd Go SDK                     │  │
│  │   - Direct gRPC API calls               │  │
│  │   - Fine-grained control                │  │
│  │   - Custom OCI specs                    │  │
│  │   - Advanced networking                 │  │
│  └─────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

### Build Operations (CLI-based)

**File**: `builders/image_builder.go`

```go
// Using CLI commands
func (b *ImageBuilder) Build(buildArgs map[string]string, target string) error {
    args := []string{
        "nerdctl",
        "--profile", b.profile,
        "--",
        "--namespace", b.namespace,
        "build",
        "-f", b.dockerfile,
        "-t", b.imageName,
        "--target", target,
        b.projectPath,
    }
    
    cmd := exec.Command("colima", args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

**Advantages:**
- ✅ Simple implementation
- ✅ Leverages BuildKit features
- ✅ User-friendly progress output
- ✅ Proven, stable tooling
- ✅ No complex API learning curve

**Used for:**
- `dvm build` - Building dev container images
- Image pulling
- Image inspection

### Runtime Operations (API-based)

**File**: `operators/containerd_runtime.go`

```go
// Using containerd Go SDK
import (
    "github.com/containerd/containerd"
    "github.com/containerd/containerd/cio"
    "github.com/containerd/containerd/oci"
    "github.com/containerd/containerd/namespaces"
)

func (c *ContainerdRuntime) StartWorkspace(ctx context.Context, opts WorkspaceOptions) (string, error) {
    // Direct API calls to containerd
    ctx = namespaces.WithNamespace(ctx, c.namespace)
    
    image, err := c.client.GetImage(ctx, opts.ImageName)
    if err != nil {
        return "", err
    }
    
    // Create container with custom OCI spec
    container, err := c.client.NewContainer(
        ctx,
        opts.WorkspaceName,
        containerd.WithImage(image),
        containerd.WithNewSpec(
            oci.WithImageConfigArgs(image, command),
            oci.WithMounts(customMounts),
            oci.WithEnv(envVars),
        ),
    )
    
    // Create and start task
    task, err := container.NewTask(ctx, cio.NewCreator())
    return container.ID(), task.Start(ctx)
}
```

**Advantages:**
- ✅ Fine-grained control over OCI specs
- ✅ No subprocess overhead
- ✅ Better error handling
- ✅ Advanced features (custom mounts, networking)
- ✅ Direct access to containerd features

**Used for:**
- `dvm attach` - Attaching to containers
- Container lifecycle (start, stop, delete)
- Exec into running containers
- Custom OCI specifications

### Why This Hybrid Approach?

#### Option 1: Pure CLI (rejected)
```go
// Everything via CLI
exec.Command("colima", "nerdctl", "run", ...)    // Limited control
exec.Command("colima", "nerdctl", "exec", ...)   // No custom specs
exec.Command("colima", "nerdctl", "attach", ...) // Less flexible
```
❌ Limited OCI spec control
❌ Harder to customize
❌ Subprocess overhead for everything

#### Option 2: Pure API (overkill for MVP)
```go
// Everything via API
buildkit.Build(...)  // Complex BuildKit API
containerd.Run(...)  // Good
containerd.Exec(...) // Good
```
❌ BuildKit API is complex
❌ More code to maintain
✅ Full control everywhere

#### Option 3: Hybrid (chosen) ✅
```go
// Build via CLI
exec.Command("nerdctl", "build", ...)  // Simple, works

// Runtime via API
containerd.NewContainer(...)           // Full control
```
✅ Best of both worlds
✅ Simple where possible
✅ Powerful where needed
✅ Easy to extend

### Runtime Interface

Both approaches implement the same interface:

```go
// operators/runtime_interface.go
type ContainerRuntime interface {
    StartWorkspace(ctx context.Context, opts WorkspaceOptions) (string, error)
    StopWorkspace(ctx context.Context, workspaceID string) error
    AttachToWorkspace(ctx context.Context, workspaceID string) error
    DeleteWorkspace(ctx context.Context, workspaceID string) error
    GetRuntimeType() string
}
```

### Future Extensions

**Build Methods** (configurable):
```yaml
# workspace.yaml
spec:
  build:
    method: cli  # Default (current)
    # Future options:
    # method: api        # Use BuildKit API directly
    # method: kaniko     # For Kubernetes
    # method: buildah    # For rootless
```

**Runtime Options**:
```yaml
spec:
  runtime:
    type: containerd  # Current
    # Future options:
    # type: docker
    # type: podman
    # type: kubernetes
```

### Testing Strategy

**CLI Build Testing**:
```go
func TestImageBuilder(t *testing.T) {
    builder := NewImageBuilder("local-med", "devopsmaestro", "/path", "image:tag", "Dockerfile")
    
    // Can mock exec.Command for testing
    err := builder.Build(nil, "dev")
    assert.NoError(t, err)
}
```

**API Runtime Testing**:
```go
func TestContainerdRuntime(t *testing.T) {
    // Can use containerd's test helpers
    runtime := NewContainerdRuntime()
    
    id, err := runtime.StartWorkspace(ctx, opts)
    assert.NoError(t, err)
    assert.NotEmpty(t, id)
}
```

### Performance Comparison

| Operation | CLI | API | Winner |
|-----------|-----|-----|--------|
| Build image | Fast (nerdctl/BuildKit) | Complex (BuildKit API) | **CLI** |
| Start container | Slower (subprocess) | Fast (direct gRPC) | **API** |
| Attach to container | Limited (TTY handling) | Full control | **API** |
| Exec command | Overhead (new process) | Efficient (reuse conn) | **API** |
| Pull image | Good (progress bars) | Good (but need to implement UI) | **CLI** |

### Conclusion

The hybrid approach gives us:
- **Simplicity** for builds (CLI)
- **Power** for runtime (API)
- **Flexibility** to change either independently
- **Best user experience** for both operations

This aligns with the MVP goal: **Get it working quickly with room to grow**.
