package builders

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	bkclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
	"github.com/moby/buildkit/util/progress/progressui"
)

// BuildKitBuilder builds container images using BuildKit gRPC API.
// Works with: Colima (containerd mode)
//
// This builder uses direct API connections to containerd and BuildKit,
// providing better performance and control compared to CLI-based builders.
type BuildKitBuilder struct {
	containerdClient *client.Client
	buildkitClient   *bkclient.Client
	namespace        string
	projectPath      string
	imageName        string
	dockerfile       string
	containerdSocket string
	buildkitSocket   string
}

// NewBuildKitBuilder creates a new BuildKit API-based image builder.
func NewBuildKitBuilder(cfg BuilderConfig) (*BuildKitBuilder, error) {
	// Get socket paths from platform
	containerdSocket := cfg.Platform.GetContainerdSocket()
	buildkitSocket := cfg.Platform.GetBuildKitSocket()

	if containerdSocket == "" {
		return nil, fmt.Errorf("platform %s does not provide containerd socket", cfg.Platform.Type)
	}
	if buildkitSocket == "" {
		return nil, fmt.Errorf("platform %s does not provide buildkit socket", cfg.Platform.Type)
	}

	// Connect to containerd for image operations
	containerdClient, err := client.New(containerdSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd at %s: %w", containerdSocket, err)
	}

	// Connect to BuildKit for build operations
	buildkitClient, err := bkclient.New(context.Background(), "unix://"+buildkitSocket)
	if err != nil {
		containerdClient.Close()
		return nil, fmt.Errorf("failed to connect to BuildKit at %s: %w", buildkitSocket, err)
	}

	return &BuildKitBuilder{
		containerdClient: containerdClient,
		buildkitClient:   buildkitClient,
		namespace:        cfg.Namespace,
		projectPath:      cfg.ProjectPath,
		imageName:        cfg.ImageName,
		dockerfile:       cfg.Dockerfile,
		containerdSocket: containerdSocket,
		buildkitSocket:   buildkitSocket,
	}, nil
}

// Build builds the container image using BuildKit gRPC API.
func (b *BuildKitBuilder) Build(ctx context.Context, opts BuildOptions) error {
	fmt.Printf("Building image: %s\n", b.imageName)
	fmt.Printf("Using BuildKit gRPC API\n")
	fmt.Printf("BuildKit socket: %s\n", b.buildkitSocket)
	fmt.Printf("Namespace: %s\n\n", b.namespace)

	// Determine dockerfile path
	dockerfilePath := b.dockerfile
	if dockerfilePath == "" {
		dockerfilePath = filepath.Join(b.projectPath, "Dockerfile")
	}

	// Prepare build options
	// Note: We use "oci" export type to save image to a file, then import to containerd.
	// This is more reliable than trying to use the "image" type which only stores in BuildKit cache.
	solveOpts := bkclient.SolveOpt{
		LocalDirs: map[string]string{
			"context":    b.projectPath,
			"dockerfile": filepath.Dir(dockerfilePath),
		},
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"filename": filepath.Base(dockerfilePath),
		},
		// Export directly to containerd image store
		Exports: []bkclient.ExportEntry{
			{
				Type: bkclient.ExporterImage,
				Attrs: map[string]string{
					"name":                 b.imageName,
					"name-canonical":       "true",
					"push":                 "false",
					"containerd.store":     "true", // Store directly in containerd
					"containerd.namespace": b.namespace,
				},
			},
		},
	}

	// Add target if specified
	if opts.Target != "" {
		solveOpts.FrontendAttrs["target"] = opts.Target
	}

	// Add no-cache if specified
	if opts.NoCache {
		solveOpts.FrontendAttrs["no-cache"] = ""
	}

	// Add build args
	for key, value := range opts.BuildArgs {
		solveOpts.FrontendAttrs["build-arg:"+key] = value
	}

	// Create session for build context
	s, err := session.NewSession(ctx, "dvm-build")
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Add auth provider for private registries
	authProvider := authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{})
	s.Allow(authProvider)

	// Add SSH agent forwarding for git operations with private repos
	if sshAuthSock := os.Getenv("SSH_AUTH_SOCK"); sshAuthSock != "" {
		sshProvider, err := sshprovider.NewSSHAgentProvider([]sshprovider.AgentConfig{
			{
				ID:    "default",
				Paths: []string{sshAuthSock},
			},
		})
		if err == nil {
			s.Allow(sshProvider)
			fmt.Println("  SSH agent forwarding enabled")
		}
	}

	// Create progress display
	displayCh := make(chan *bkclient.SolveStatus)
	var wg sync.WaitGroup
	var displayErr, solveErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		display, err := progressui.NewDisplay(os.Stdout, progressui.PlainMode)
		if err != nil {
			displayErr = err
			return
		}
		_, displayErr = display.UpdateFrom(ctx, displayCh)
	}()

	// Run build (Solve closes displayCh when complete)
	_, solveErr = b.buildkitClient.Solve(ctx, nil, solveOpts, displayCh)

	wg.Wait()

	if displayErr != nil {
		return fmt.Errorf("progress display error: %w", displayErr)
	}
	if solveErr != nil {
		return fmt.Errorf("build failed: %w", solveErr)
	}

	fmt.Printf("\n✓ Image built successfully: %s\n", b.imageName)
	return nil
}

// ImageExists checks if an image already exists using containerd API.
func (b *BuildKitBuilder) ImageExists(ctx context.Context) (bool, error) {
	ctx = namespaces.WithNamespace(ctx, b.namespace)

	_, err := b.containerdClient.ImageService().Get(ctx, b.imageName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Close releases all client connections.
func (b *BuildKitBuilder) Close() error {
	var errs []error

	if b.buildkitClient != nil {
		if err := b.buildkitClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("buildkit close: %w", err))
		}
	}

	if b.containerdClient != nil {
		if err := b.containerdClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("containerd close: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}
	return nil
}
