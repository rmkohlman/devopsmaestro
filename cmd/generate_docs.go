package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// generateDocsCmd generates man pages and markdown docs for the dvm CLI.
// It is hidden from normal users — this is a developer/release tool.
var generateDocsCmd = &cobra.Command{
	Use:    "generate-docs",
	Short:  "Generate documentation (man pages, markdown) for dvm",
	Hidden: true,
	Long: `Generate documentation files from dvm's command definitions.

Outputs man pages (section 1) and/or markdown reference docs for dvm and all
subcommands. Intended for use during releases — not surfaced in normal help.

Examples:
  dvm generate-docs --man-pages --output-dir ./docs/man/
  dvm generate-docs --markdown --output-dir ./docs/reference/
  dvm generate-docs --man-pages --markdown --output-dir ./dist/docs/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir, _ := cmd.Flags().GetString("output-dir")
		manPages, _ := cmd.Flags().GetBool("man-pages")
		markdown, _ := cmd.Flags().GetBool("markdown")

		if outputDir == "" {
			outputDir = "./docs/man"
		}

		if !manPages && !markdown {
			return fmt.Errorf("specify at least one output format: --man-pages and/or --markdown")
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
		}

		if manPages {
			if err := generateManPages(rootCmd, outputDir); err != nil {
				return fmt.Errorf("man page generation failed: %w", err)
			}
			fmt.Fprintf(os.Stdout, "Man pages written to %s\n", outputDir)
		}

		if markdown {
			if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
				return fmt.Errorf("markdown generation failed: %w", err)
			}
			fmt.Fprintf(os.Stdout, "Markdown docs written to %s\n", outputDir)
		}

		return nil
	},
}

// generateManPages writes section-1 man pages for cmd and all subcommands.
func generateManPages(root *cobra.Command, dir string) error {
	now := time.Now()
	header := &doc.GenManHeader{
		Title:   "DVM",
		Section: "1",
		Date:    &now,
		Source:  "DevOpsMaestro",
		Manual:  "DevOpsMaestro Manual",
	}
	return doc.GenManTree(root, header, dir)
}

func init() {
	generateDocsCmd.Flags().String("output-dir", "./docs/man", "Directory to write generated documentation")
	generateDocsCmd.Flags().Bool("man-pages", false, "Generate man pages (section 1)")
	generateDocsCmd.Flags().Bool("markdown", false, "Generate markdown reference docs")
	rootCmd.AddCommand(generateDocsCmd)
}
