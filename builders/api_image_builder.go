package builders

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	bkclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
	"github.com/moby/buildkit/util/progress/progressui"
)

// APIImageBuilder builds container images using BuildKit gRPC API
type APIImageBuilder struct {
	containerdClient *client.Client
	buildkitClient   *bkclient.Client
	namespace        string
	projectPath      string
	imageName        string
	dockerfile       string
	containerdSocket string
	buildkitSocket   string
}

// NewAPIImageBuilder creates a new API-based image builder
func NewAPIImageBuilder(containerdSocket, buildkitSocket, namespace, projectPath, imageName, dockerfile string) (*APIImageBuilder, error) {
	// Connect to containerd for image operations
	containerdClient, err := client.New(containerdSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	// Connect to BuildKit on separate socket
	buildkitClient, err := bkclient.New(context.Background(), "unix://"+buildkitSocket)
	if err != nil {
		containerdClient.Close()
		return nil, fmt.Errorf("failed to connect to BuildKit: %w", err)
	}

	return &APIImageBuilder{
		containerdClient: containerdClient,
		buildkitClient:   buildkitClient,
		namespace:        namespace,
		projectPath:      projectPath,
		imageName:        imageName,
		dockerfile:       dockerfile,
		containerdSocket: containerdSocket,
		buildkitSocket:   buildkitSocket,
	}, nil
}

// Build builds the container image using BuildKit gRPC API
func (b *APIImageBuilder) Build(ctx context.Context, buildArgs map[string]string, target string) error {
	fmt.Printf("Building image: %s\n", b.imageName)
	fmt.Printf("Using BuildKit gRPC API\n")
	fmt.Printf("BuildKit socket: %s\n", b.buildkitSocket)
	fmt.Printf("Namespace: %s\n\n", b.namespace)

	// Prepare build options
	opts := bkclient.SolveOpt{
		LocalDirs: map[string]string{
			"context":    b.projectPath,
			"dockerfile": filepath.Dir(b.dockerfile),
		},
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"filename": filepath.Base(b.dockerfile),
			"target":   target,
		},
		Exports: []bkclient.ExportEntry{
			{
				Type: "image",
				Attrs: map[string]string{
					"name": b.imageName,
					"push": "false",
				},
			},
		},
	}

	// Add build args
	if len(buildArgs) > 0 {
		for key, value := range buildArgs {
			opts.FrontendAttrs["build-arg:"+key] = value
		}
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
	// This uses the host's SSH agent (if available) or SSH keys from ~/.ssh
	sshProvider, err := sshprovider.NewSSHAgentProvider([]sshprovider.AgentConfig{
		{
			ID:    "default",
			Paths: []string{os.Getenv("SSH_AUTH_SOCK")},
		},
	})
	if err == nil {
		s.Allow(sshProvider)
		fmt.Println("  SSH agent forwarding enabled")
	}

	// Create progress display
	displayCh := make(chan *bkclient.SolveStatus)
	var wg sync.WaitGroup
	var displayErr, solveErr error

	wg.Add(1)

	// Display progress
	go func() {
		defer wg.Done()
		display, err := progressui.NewDisplay(os.Stdout, progressui.PlainMode)
		if err != nil {
			displayErr = err
			return
		}
		_, displayErr = display.UpdateFrom(ctx, displayCh)
	}()

	// Run build (synchronously in main goroutine)
	// Note: Solve closes the displayCh when complete
	_, solveErr = b.buildkitClient.Solve(ctx, nil, opts, displayCh)

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

// ImageExists checks if an image already exists using containerd API
func (b *APIImageBuilder) ImageExists(ctx context.Context) (bool, error) {
	ctx = namespaces.WithNamespace(ctx, b.namespace)

	_, err := b.containerdClient.ImageService().Get(ctx, b.imageName)
	if err != nil {
		// Check if error is "not found" error
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ListImages lists all images in the namespace using containerd API
func (b *APIImageBuilder) ListImages(ctx context.Context) ([]images.Image, error) {
	ctx = namespaces.WithNamespace(ctx, b.namespace)
	return b.containerdClient.ImageService().List(ctx)
}

// PullImage pulls an image using BuildKit
func (b *APIImageBuilder) PullImage(ctx context.Context, imageName string) error {
	fmt.Printf("Pulling image: %s\n", imageName)

	opts := bkclient.SolveOpt{
		Frontend: "gateway.v0",
		FrontendAttrs: map[string]string{
			"source": imageName,
		},
		Exports: []bkclient.ExportEntry{
			{
				Type: "image",
				Attrs: map[string]string{
					"name": imageName,
					"push": "false",
				},
			},
		},
	}

	_, err := b.buildkitClient.Solve(ctx, nil, opts, nil)
	return err
}

// Close closes all client connections
func (b *APIImageBuilder) Close() error {
	var errs []error

	if b.buildkitClient != nil {
		if err := b.buildkitClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if b.containerdClient != nil {
		if err := b.containerdClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}
	return nil
}

// errGroup is a simple error group implementation
type errGroup struct {
	errs []error
	fns  []func() error
}

func (eg *errGroup) Go(fn func() error) {
	eg.fns = append(eg.fns, fn)
}

func (eg *errGroup) Wait() error {
	errCh := make(chan error, len(eg.fns))

	for _, fn := range eg.fns {
		go func(f func() error) {
			errCh <- f()
		}(fn)
	}

	for range eg.fns {
		if err := <-errCh; err != nil {
			eg.errs = append(eg.errs, err)
		}
	}

	if len(eg.errs) > 0 {
		return eg.errs[0]
	}
	return nil
}
