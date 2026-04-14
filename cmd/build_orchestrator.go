package cmd

import (
	"context"
	"database/sql"
	"devopsmaestro/config"
	"devopsmaestro/models"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func buildWorkspace(cmd *cobra.Command) error {
	slog.Info("starting build")

	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Create timeout context from flag
	ctx := context.Background()
	if buildTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, buildTimeout)
		defer cancel()
	}

	bc := &buildContext{
		ds:     sqlDS,
		ctx:    ctx,
		output: os.Stdout,
	}

	// Phase 1: Resolve target workspace
	if err := bc.resolveWorkspaceTarget(); err != nil {
		return err
	}

	// Dry-run: preview what would be built
	if buildDryRun {
		bc.renderPlain(fmt.Sprintf("Would build image for workspace %q in app %q", bc.workspaceName, bc.appName))
		if buildNocache {
			bc.renderPlain("  --no-cache: would skip registry cache")
		}
		if buildPush {
			bc.renderPlain("  --push: would push image to local registry")
		}
		if buildTarget != "dev" {
			bc.renderPlain(fmt.Sprintf("  --target: %s", buildTarget))
		}
		return nil
	}

	bc.renderInfof("Building workspace: %s/%s", bc.appName, bc.workspaceName)
	bc.renderInfof("App path: %s", bc.app.Path)
	bc.renderBlank()
	slog.Debug("app details", "path", bc.app.Path, "id", bc.app.ID)

	// --- Session persistence: create session for single-workspace build ---
	// This ensures `dvm build status` always reflects the latest build attempt,
	// whether it was a single-workspace or parallel build.
	sessionID := uuid.New().String()
	buildStart := time.Now().UTC()
	cleanupBuildSessions(sqlDS)

	session := &models.BuildSession{
		ID:              sessionID,
		StartedAt:       buildStart,
		Status:          "running",
		TotalWorkspaces: 1,
	}
	if err := sqlDS.CreateBuildSession(session); err != nil {
		slog.Warn("failed to create build session", "error", err)
	}

	bsw := &models.BuildSessionWorkspace{
		SessionID:   sessionID,
		WorkspaceID: bc.workspace.ID,
		Status:      "building",
		StartedAt:   sql.NullTime{Time: buildStart, Valid: true},
	}
	if err := sqlDS.CreateBuildSessionWorkspace(bsw); err != nil {
		slog.Warn("failed to create build session workspace entry", "error", err)
	}
	// Read back the assigned ID for later updates
	entries, _ := sqlDS.GetBuildSessionWorkspaces(sessionID)
	var bswID int
	if len(entries) > 0 {
		bswID = entries[0].ID
	}

	// finalizeBuildSession updates the session and workspace entry on exit.
	// Uses named return so deferred func captures final buildErr.
	var buildErr error
	defer func() {
		completedAt := time.Now().UTC()
		duration := int64(completedAt.Sub(buildStart).Seconds())

		wsStatus := "succeeded"
		var errMsg string
		if buildErr != nil {
			wsStatus = "failed"
			errMsg = buildErr.Error()
		}

		if bswID > 0 {
			upd := &models.BuildSessionWorkspace{
				ID:              bswID,
				SessionID:       sessionID,
				WorkspaceID:     bc.workspace.ID,
				Status:          wsStatus,
				StartedAt:       sql.NullTime{Time: buildStart, Valid: true},
				CompletedAt:     sql.NullTime{Time: completedAt, Valid: true},
				DurationSeconds: sql.NullInt64{Int64: duration, Valid: true},
				ImageTag:        sql.NullString{String: bc.imageName, Valid: bc.imageName != ""},
				ErrorMessage:    sql.NullString{String: errMsg, Valid: errMsg != ""},
			}
			if err := sqlDS.UpdateBuildSessionWorkspace(upd); err != nil {
				slog.Warn("failed to update build session workspace", "error", err)
			}
		}

		sessStatus := "completed"
		succeeded, failed := 1, 0
		if buildErr != nil {
			sessStatus = "failed"
			succeeded, failed = 0, 1
		}
		sess := &models.BuildSession{
			ID:              sessionID,
			StartedAt:       buildStart,
			CompletedAt:     sql.NullTime{Time: completedAt, Valid: true},
			Status:          sessStatus,
			TotalWorkspaces: 1,
			Succeeded:       succeeded,
			Failed:          failed,
		}
		if err := sqlDS.UpdateBuildSession(sess); err != nil {
			slog.Warn("failed to update build session", "error", err)
		}
	}()

	if err := bc.validateAppPath(); err != nil {
		buildErr = err
		return buildErr
	}

	// Phase 2: Platform & registry
	if err := bc.detectBuildPlatform(); err != nil {
		buildErr = err
		return buildErr
	}
	if err := bc.prepareRegistry(); err != nil {
		buildErr = err
		return buildErr
	}

	// Phase 3: Dockerfile detection & workspace spec
	bc.checkDockerfile()
	if err := bc.prepareWorkspaceSpec(); err != nil {
		buildErr = err
		return buildErr
	}

	// Phase 4: Source, staging, language detection
	if err := bc.prepareSourceAndStaging(); err != nil {
		buildErr = err
		return buildErr
	}
	defer func() {
		if err := os.RemoveAll(bc.stagingDir); err != nil {
			slog.Warn("failed to clean up staging directory", "path", bc.stagingDir, "error", err)
		} else {
			slog.Debug("cleaned up staging directory", "path", bc.stagingDir)
		}
	}()

	// Phase 5: CA certs & nvim config
	if err := bc.resolveCACerts(); err != nil {
		buildErr = err
		return buildErr
	}
	if err := bc.generateNvimConfiguration(); err != nil {
		buildErr = err
		return buildErr
	}

	// Phase 6: Dockerfile generation & build
	if err := bc.generateDockerfileAndResolveArgs(); err != nil {
		buildErr = err
		return buildErr
	}

	// Phase 6b: Validate staging directory (warn on missing COPY sources)
	if err := bc.validateStagingDirectory(); err != nil {
		buildErr = err
		return buildErr
	}

	skipped, err := bc.buildImage()
	if bc.builder != nil {
		defer bc.builder.Close()
	}
	if err != nil {
		buildErr = err
		return buildErr
	}
	if skipped {
		return nil
	}

	// Phase 7: Post-build (DB update, registry push, summary)
	bc.postBuild()

	return nil
}

// prepareCACerts resolves CA certificates from MaestroVault and writes them
// to the staging directory's certs/ subdirectory. The Dockerfile generator
// will COPY these into the image and update the system trust store.
// All errors are fatal — a missing or invalid cert should fail the build.
func prepareCACerts(stagingDir string, caCerts []models.CACertConfig) error {
	// Validate cert configs
	if err := models.ValidateCACerts(caCerts); err != nil {
		return fmt.Errorf("invalid CA certificate configuration: %w", err)
	}

	// Resolve vault token (fatal if missing — certs require vault)
	token, tokenErr := config.ResolveVaultToken()
	if tokenErr != nil || token == "" {
		return fmt.Errorf("CA certificates require MaestroVault but no vault token is configured. Hint: Configure vault with: dvm admin vault configure")
	}
	if err := config.EnsureVaultDaemon(); err != nil {
		return fmt.Errorf("failed to start vault daemon for CA cert resolution: %w", err)
	}
	vb, err := config.NewVaultBackend(token)
	if err != nil {
		return fmt.Errorf("failed to create vault backend for CA cert resolution: %w", err)
	}

	return prepareCACertsWithBackend(stagingDir, caCerts, vb)
}

// prepareCACertsWithBackend is the injectable core of prepareCACerts. It
// accepts an already-constructed SecretBackend so tests can supply a mock
// without needing a live MaestroVault daemon.
func prepareCACertsWithBackend(stagingDir string, caCerts []models.CACertConfig, backend config.SecretBackend) error {
	// Create certs directory in staging
	certsDir := filepath.Join(stagingDir, "certs")
	if err := os.MkdirAll(certsDir, 0700); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Resolve each cert from vault and write to staging
	for _, cert := range caCerts {
		var pemContent string
		var err error

		// Check if this is a field-level request
		if cert.VaultField != "" {
			// Use GetField for field-level access
			if fb, ok := backend.(config.FieldCapableBackend); ok {
				pemContent, err = fb.GetField(cert.VaultSecret, cert.VaultEnvironment, cert.VaultField)
			} else {
				return fmt.Errorf("vault backend does not support field-level access for cert %q", cert.Name)
			}
		} else {
			pemContent, err = backend.Get(cert.VaultSecret, cert.VaultEnvironment)
		}
		if err != nil {
			return fmt.Errorf("failed to resolve CA certificate %q from vault: %w", cert.Name, err)
		}

		slog.Debug("raw PEM content from vault",
			"name", cert.Name,
			"length", len(pemContent),
			"first80", fmt.Sprintf("%q", pemContent[:min(80, len(pemContent))]),
			"last40", fmt.Sprintf("%q", pemContent[max(0, len(pemContent)-40):]),
		)

		// Normalize PEM content — vault backends may collapse newlines to spaces
		pemContent = models.NormalizePEMContent(pemContent)

		// Validate PEM content
		if err := models.ValidatePEMContent(pemContent); err != nil {
			return fmt.Errorf("CA certificate %q has invalid content: %w", cert.Name, err)
		}

		// Path traversal defense: ensure name is just a filename
		safeName := filepath.Base(cert.Name)
		if safeName != cert.Name {
			return fmt.Errorf("CA certificate name %q contains path separators", cert.Name)
		}

		certPath := filepath.Join(certsDir, safeName+".crt")

		// Verify the resolved path is within certsDir (defense in depth)
		cleanPath := filepath.Clean(certPath)
		if !strings.HasPrefix(cleanPath, filepath.Clean(certsDir)) {
			return fmt.Errorf("CA certificate path %q escapes certs directory", cert.Name)
		}

		if err := os.WriteFile(certPath, []byte(pemContent), 0600); err != nil {
			return fmt.Errorf("failed to write CA certificate %q: %w", cert.Name, err)
		}

		slog.Debug("wrote CA certificate", "name", cert.Name, "path", certPath)
	}

	return nil
}
