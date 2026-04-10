package cmd

import (
	"fmt"

	"devopsmaestro/pkg/mirror"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// runGetBranches implements the get branches command.
func runGetBranches(cmd *cobra.Command, _ []string) error {
	repoName, _ := cmd.Flags().GetString("repo")
	if repoName == "" {
		return fmt.Errorf("--repo flag is required")
	}

	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	repo, err := dataStore.GetGitRepoByName(repoName)
	if err != nil {
		return fmt.Errorf("gitrepo '%s' not found", repoName)
	}

	mm := getMirrorManager(cmd)
	inspector, ok := mm.(mirror.MirrorInspector)
	if !ok {
		return fmt.Errorf("mirror manager does not support inspection")
	}

	branches, err := inspector.ListBranches(repo.Slug)
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	format, _ := cmd.Flags().GetString("output")
	return renderRefs(cmd, branches, format, "BRANCH")
}

// runGetTags implements the get tags command.
func runGetTags(cmd *cobra.Command, _ []string) error {
	repoName, _ := cmd.Flags().GetString("repo")
	if repoName == "" {
		return fmt.Errorf("--repo flag is required")
	}

	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	repo, err := dataStore.GetGitRepoByName(repoName)
	if err != nil {
		return fmt.Errorf("gitrepo '%s' not found", repoName)
	}

	mm := getMirrorManager(cmd)
	inspector, ok := mm.(mirror.MirrorInspector)
	if !ok {
		return fmt.Errorf("mirror manager does not support inspection")
	}

	tags, err := inspector.ListTags(repo.Slug)
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	format, _ := cmd.Flags().GetString("output")
	return renderRefs(cmd, tags, format, "TAG")
}

// renderRefs renders a slice of RefInfo as a table or structured output.
func renderRefs(cmd *cobra.Command, refs []mirror.RefInfo, format, header string) error {
	rows := make([][]string, len(refs))
	for i, ref := range refs {
		rows[i] = []string{ref.Name, ref.Hash, ref.Date}
	}

	tableData := render.TableData{
		Headers: []string{header, "COMMIT", "DATE"},
		Rows:    rows,
	}

	return render.OutputTo(cmd.OutOrStdout(), format, tableData, render.Options{
		Type: render.TypeTable,
	})
}
