package handlers

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

const KindGitRepo = "GitRepo"

// GitRepoHandler handles GitRepo resources.
type GitRepoHandler struct{}

// NewGitRepoHandler creates a new GitRepo handler.
func NewGitRepoHandler() *GitRepoHandler {
	return &GitRepoHandler{}
}

func (h *GitRepoHandler) Kind() string {
	return KindGitRepo
}

// Apply creates or updates a git repo from YAML data.
func (h *GitRepoHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	var gitRepoYAML models.GitRepoYAML
	if err := yaml.Unmarshal(data, &gitRepoYAML); err != nil {
		return nil, fmt.Errorf("failed to parse git repo YAML: %w", err)
	}

	// Convert to model
	repo := &models.GitRepoDB{}
	repo.FromYAML(gitRepoYAML)

	// Get the datastore
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Resolve Credential name → CredentialID if specified in YAML
	credentialName := gitRepoYAML.Spec.Credential
	if credentialName != "" {
		cred, credErr := ds.GetCredentialByName(credentialName)
		if credErr != nil {
			return nil, fmt.Errorf("credential '%s' not found: %w", credentialName, credErr)
		}
		repo.CredentialID = sql.NullInt64{Int64: cred.ID, Valid: true}
	}

	// Check if git repo exists
	existing, _ := ds.GetGitRepoByName(repo.Name)
	if existing != nil {
		// Update existing
		repo.ID = existing.ID
		repo.Slug = existing.Slug
		if err := ds.UpdateGitRepo(repo); err != nil {
			return nil, fmt.Errorf("failed to update git repo: %w", err)
		}
	} else {
		// Generate slug from URL
		if repo.URL != "" {
			slug, slugErr := mirror.GenerateSlug(repo.URL)
			if slugErr == nil {
				repo.Slug = slug
			}
		}

		// Create new
		if err := ds.CreateGitRepo(repo); err != nil {
			return nil, fmt.Errorf("failed to create git repo: %w", err)
		}
		// Fetch to get the ID
		repo, err = ds.GetGitRepoByName(repo.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created git repo: %w", err)
		}
	}

	return &GitRepoResource{gitRepo: repo, credentialName: credentialName}, nil
}

// Get retrieves a git repo by name.
func (h *GitRepoHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	repo, err := ds.GetGitRepoByName(name)
	if err != nil {
		return nil, err
	}

	// Resolve CredentialID → credential name
	credentialName := ""
	if repo.CredentialID.Valid {
		creds, credErr := ds.ListAllCredentials()
		if credErr == nil {
			for _, c := range creds {
				if c.ID == repo.CredentialID.Int64 {
					credentialName = c.Name
					break
				}
			}
		}
	}

	return &GitRepoResource{gitRepo: repo, credentialName: credentialName}, nil
}

// List returns all git repos.
func (h *GitRepoHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	repos, err := ds.ListGitRepos()
	if err != nil {
		return nil, err
	}

	// Build credential ID → name lookup for resolving CredentialID references
	credNames := make(map[int64]string)
	creds, credErr := ds.ListAllCredentials()
	if credErr == nil {
		for _, c := range creds {
			credNames[c.ID] = c.Name
		}
	}

	result := make([]resource.Resource, len(repos))
	for i := range repos {
		credentialName := ""
		if repos[i].CredentialID.Valid {
			credentialName = credNames[repos[i].CredentialID.Int64]
		}
		result[i] = &GitRepoResource{gitRepo: &repos[i], credentialName: credentialName}
	}
	return result, nil
}

// Delete removes a git repo by name.
func (h *GitRepoHandler) Delete(ctx resource.Context, name string) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return err
	}

	// Check existence at handler level (consistent with other handlers)
	_, err = ds.GetGitRepoByName(name)
	if err != nil {
		return err
	}

	return ds.DeleteGitRepo(name)
}

// ToYAML serializes a git repo to YAML.
func (h *GitRepoHandler) ToYAML(res resource.Resource) ([]byte, error) {
	gr, ok := res.(*GitRepoResource)
	if !ok {
		return nil, fmt.Errorf("expected GitRepoResource, got %T", res)
	}

	yamlDoc := gr.gitRepo.ToYAML(gr.credentialName)
	return yaml.Marshal(yamlDoc)
}

// GitRepoResource wraps a models.GitRepoDB to implement resource.Resource.
type GitRepoResource struct {
	gitRepo        *models.GitRepoDB
	credentialName string // Name of the associated credential, if any
}

// NewGitRepoResource creates a new GitRepoResource from a model.
// credentialName is optional — pass as extra[0] if available.
func NewGitRepoResource(repo *models.GitRepoDB, extra ...string) *GitRepoResource {
	credentialName := ""
	if len(extra) > 0 {
		credentialName = extra[0]
	}
	return &GitRepoResource{gitRepo: repo, credentialName: credentialName}
}

func (r *GitRepoResource) GetKind() string {
	return KindGitRepo
}

func (r *GitRepoResource) GetName() string {
	return r.gitRepo.Name
}

func (r *GitRepoResource) Validate() error {
	if r.gitRepo.Name == "" {
		return fmt.Errorf("git repo name is required")
	}
	if r.gitRepo.URL == "" {
		return fmt.Errorf("git repo URL is required")
	}
	return nil
}

// GitRepo returns the underlying models.GitRepoDB.
func (r *GitRepoResource) GitRepo() *models.GitRepoDB {
	return r.gitRepo
}
