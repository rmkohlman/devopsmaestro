package cmd

import (
	"fmt"

	"devopsmaestro/db"
)

// updateGitRepoFields applies non-interactive field patches to a gitrepo.
// Empty string values are skipped (field is not updated).
// This is the underlying helper called by the edit command's non-interactive path.
func updateGitRepoFields(ds db.DataStore, name, url, defaultRef, authType, credential string) error {
	repo, err := ds.GetGitRepoByName(name)
	if err != nil {
		return fmt.Errorf("gitrepo '%s' not found: %w", name, err)
	}

	// Apply non-empty field patches
	if url != "" {
		repo.URL = url
	}
	if defaultRef != "" {
		repo.DefaultRef = defaultRef
	}
	if authType != "" {
		repo.AuthType = authType
	}
	if credential != "" {
		// Look up credential by name and set ID
		cred, credErr := ds.GetCredentialByName(credential)
		if credErr != nil {
			return fmt.Errorf("credential '%s' not found: %w", credential, credErr)
		}
		repo.CredentialID.Int64 = cred.ID
		repo.CredentialID.Valid = true
	}

	if err := ds.UpdateGitRepo(repo); err != nil {
		return fmt.Errorf("failed to update gitrepo '%s': %w", name, err)
	}

	return nil
}
