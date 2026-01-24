package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/operators"
	"devopsmaestro/templates"
	"devopsmaestro/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	buildForce   bool
	buildNocache bool
	buildTarget  string
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build workspace container image",
	Long: `Build a development container image for the active workspace.

This command:
- Detects the project language
- Generates or extends Dockerfile with dev tools
- Builds the image using containerd (nerdctl)
- Tags as dvm-<workspace>-<project>:latest

Examples:
  dvm build
  dvm build --force
  dvm build --no-cache
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return buildWorkspace(cmd)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&buildForce, "force", false, "Force rebuild even if image exists")
	buildCmd.Flags().BoolVar(&buildNocache, "no-cache", false, "Build without using cache")
	buildCmd.Flags().StringVar(&buildTarget, "target", "dev", "Build target stage (default: dev)")
}

func buildWorkspace(cmd *cobra.Command) error {
	// Get current context
	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	projectName, err := ctxMgr.GetActiveProject()
	if err != nil {
		return fmt.Errorf("no active project set. Use 'dvm use project <name>' first")
	}

	workspaceName, err := ctxMgr.GetActiveWorkspace()
	if err != nil {
		return fmt.Errorf("no active workspace set. Use 'dvm use workspace <name>' first")
	}

	// Get datastore
	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get project
	project, err := sqlDS.GetProjectByName(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get workspace
	workspace, err := sqlDS.GetWorkspaceByName(project.ID, workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	fmt.Printf("Building workspace: %s/%s\n", projectName, workspaceName)
	fmt.Printf("Project path: %s\n\n", project.Path)

	// Verify project path exists
	if _, err := os.Stat(project.Path); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", project.Path)
	}

	// Step 1: Detect language
	fmt.Println("→ Detecting project language...")
	lang, err := utils.DetectLanguage(project.Path)
	if err != nil {
		return fmt.Errorf("failed to detect language: %w", err)
	}

	var languageName, version string
	if lang != nil {
		languageName = lang.Name
		version = utils.DetectVersion(languageName, project.Path)
		fmt.Printf("  Language: %s", languageName)
		if version != "" {
			fmt.Printf(" (version: %s)", version)
		}
		fmt.Println()
	} else {
		languageName = "unknown"
		fmt.Println("  Language: Unknown (will use generic base)")
	}

	// Step 2: Check for existing Dockerfile
	fmt.Println("\n→ Checking for Dockerfile...")
	hasDockerfile, dockerfilePath := utils.HasDockerfile(project.Path)
	if hasDockerfile {
		fmt.Printf("  Found: %s\n", dockerfilePath)
	} else {
		fmt.Println("  No Dockerfile found, will generate from scratch")
	}

	// Step 3: Generate workspace spec (for now, use defaults)
	// In the future, this will load from workspace configuration
	workspaceYAML := workspace.ToYAML(projectName)

	// Set some sensible defaults if not configured
	if workspaceYAML.Spec.Shell.Type == "" {
		workspaceYAML.Spec.Shell.Type = "zsh"
		workspaceYAML.Spec.Shell.Framework = "oh-my-zsh"
		workspaceYAML.Spec.Shell.Theme = "starship"
	}

	if workspaceYAML.Spec.Container.WorkingDir == "" {
		workspaceYAML.Spec.Container.WorkingDir = "/workspace"
	}

	// Step 4: Generate Dockerfile
	fmt.Println("\n→ Generating Dockerfile.dvm...")
	generator := builders.NewDockerfileGenerator(
		workspace,
		workspaceYAML.Spec,
		languageName,
		version,
		project.Path,
		dockerfilePath,
	)

	dockerfileContent, err := generator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Save Dockerfile
	dvmDockerfile, err := builders.SaveDockerfile(dockerfileContent, project.Path)
	if err != nil {
		return err
	}

	// Get home directory for later use
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Step 4.5: Copy nvim templates to build context if nvim config is enabled
	if workspaceYAML.Spec.Nvim.Structure != "" && workspaceYAML.Spec.Nvim.Structure != "none" {
		fmt.Println("→ Copying Neovim configuration to build context...")

		// Create .config/nvim directory in project path
		nvimConfigPath := filepath.Join(project.Path, ".config", "nvim")
		if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
			return fmt.Errorf("failed to create nvim config directory: %w", err)
		}

		// Copy templates to build context
		templatesPath := filepath.Join(homeDir, "Developer", "tools", "devopsmaestro", "templates", "nvim")

		// Copy all files from templates/nvim to project/.config/nvim
		err = filepath.Walk(templatesPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip config.yaml
			if strings.HasSuffix(path, "config.yaml") {
				return nil
			}

			// Get relative path
			relPath, err := filepath.Rel(templatesPath, path)
			if err != nil {
				return err
			}

			destPath := filepath.Join(nvimConfigPath, relPath)

			if info.IsDir() {
				return os.MkdirAll(destPath, info.Mode())
			}

			// Copy file
			input, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			return os.WriteFile(destPath, input, info.Mode())
		})

		if err != nil {
			return fmt.Errorf("failed to copy nvim templates: %w", err)
		}

		fmt.Println("✓ Neovim configuration copied")

		// Step 4.6: Generate plugin files from database
		if len(workspaceYAML.Spec.Nvim.Plugins) > 0 {
			fmt.Printf("→ Loading %d plugins from database...\n", len(workspaceYAML.Spec.Nvim.Plugins))

			pluginManager := templates.NewDBPluginManager(sqlDS)
			if err := pluginManager.GenerateLuaFilesForWorkspace(
				workspaceYAML.Spec.Nvim.Plugins,
				nvimConfigPath,
			); err != nil {
				return fmt.Errorf("failed to generate plugins: %w", err)
			}

			fmt.Printf("✓ Generated %d plugins\n", len(workspaceYAML.Spec.Nvim.Plugins))
		}
	}

	// Step 4.7: Copy shell configuration files to build context
	if workspaceYAML.Spec.Shell.Theme != "none" {
		fmt.Println("→ Copying shell configuration to build context...")

		// Copy user's shell config files if they exist
		shellFiles := []string{".zshrc", ".p10k.zsh"}
		copiedCount := 0

		for _, file := range shellFiles {
			srcPath := filepath.Join(homeDir, file)
			if _, err := os.Stat(srcPath); err == nil {
				destPath := filepath.Join(project.Path, file)
				input, err := os.ReadFile(srcPath)
				if err != nil {
					fmt.Printf("  ⚠️  Warning: Failed to read %s: %v\n", file, err)
					continue
				}

				if err := os.WriteFile(destPath, input, 0644); err != nil {
					fmt.Printf("  ⚠️  Warning: Failed to copy %s: %v\n", file, err)
					continue
				}

				copiedCount++
				fmt.Printf("  ✓ Copied %s (%d bytes)\n", file, len(input))
			} else {
				fmt.Printf("  ℹ️  %s not found, skipping\n", file)
			}
		}

		if copiedCount > 0 {
			fmt.Printf("✓ Shell configuration copied (%d files)\n", copiedCount)
		} else {
			fmt.Println("  ℹ️  No shell config files found, will use default oh-my-zsh setup")
		}
	}

	// Step 5: Build image
	imageName := fmt.Sprintf("dvm-%s-%s:latest", workspaceName, projectName)

	fmt.Printf("\n→ Building image: %s\n", imageName)

	// Get containerd and buildkit socket paths
	profile := builders.GetColimaProfile()
	containerdSocket := filepath.Join(homeDir, ".colima", profile, "containerd.sock")
	buildkitSocket := filepath.Join(homeDir, ".colima", profile, "buildkitd.sock")

	fmt.Printf("  Using Colima profile: %s\n", profile)
	fmt.Printf("  Containerd socket: %s\n", containerdSocket)
	fmt.Printf("  BuildKit socket: %s\n", buildkitSocket)
	fmt.Printf("  Namespace: devopsmaestro\n\n")

	// Create API-based image builder
	builder, err := builders.NewAPIImageBuilder(
		containerdSocket,
		buildkitSocket,
		"devopsmaestro",
		project.Path,
		imageName,
		dvmDockerfile,
	)
	if err != nil {
		return fmt.Errorf("failed to create builder: %w", err)
	}
	defer builder.Close()

	// Check if image exists
	ctx := context.Background()
	if !buildForce {
		exists, err := builder.ImageExists(ctx)
		if err == nil && exists {
			fmt.Printf("Image already exists: %s\n", imageName)
			fmt.Println("Use --force to rebuild")
			return nil
		}
	}

	// Prepare build args (from environment and config)
	buildArgs := make(map[string]string)

	// Add common build args
	if ghUser := os.Getenv("GITHUB_USERNAME"); ghUser != "" {
		buildArgs["GITHUB_USERNAME"] = ghUser
	}
	if ghPat := os.Getenv("GITHUB_PAT"); ghPat != "" {
		buildArgs["GITHUB_PAT"] = ghPat
	}

	// Build the image
	if err := builder.Build(ctx, buildArgs, buildTarget); err != nil {
		return err
	}

	// Step 5.5: Copy image from buildkit namespace to devopsmaestro namespace
	// BuildKit creates images in its own namespace, but containerd runtime expects them in devopsmaestro
	fmt.Printf("\n→ Copying image to devopsmaestro namespace...\n")

	// Use a temporary file in the VM to avoid pipe issues
	tmpFile := fmt.Sprintf("/tmp/dvm-image-%d.tar", os.Getpid())

	// Save image to temp file
	saveCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "buildkit", "image", "save", imageName, "-o", tmpFile)
	saveCmd.Stdout = os.Stdout
	saveCmd.Stderr = os.Stderr
	if err := saveCmd.Run(); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	// Load image from temp file
	loadCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "image", "load", "-i", tmpFile)
	loadCmd.Stdout = os.Stdout
	loadCmd.Stderr = os.Stderr
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	// Clean up temp file
	cleanCmd := exec.Command("colima", "--profile", profile, "ssh", "--", "sudo", "rm", "-f", tmpFile)
	cleanCmd.Run() // Ignore errors on cleanup

	fmt.Printf("✓ Image copied to devopsmaestro namespace\n")

	// Step 6: Update workspace image name in database
	workspace.ImageName = imageName
	if err := sqlDS.UpdateWorkspace(workspace); err != nil {
		fmt.Printf("Warning: Failed to update workspace image name: %v\n", err)
	}

	// Clean up Dockerfile.dvm if build was successful
	// Actually, let's keep it for debugging
	fmt.Printf("\n✓ Build complete!\n")
	fmt.Printf("  Image: %s\n", imageName)
	fmt.Printf("  Dockerfile: %s\n", dvmDockerfile)
	fmt.Println("\nNext: Attach to your workspace with: dvm attach")

	return nil
}

// Helper function to get relative path for display
func getRelativePath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}
