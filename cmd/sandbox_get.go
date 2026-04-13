package cmd

import (
	"context"
	"devopsmaestro/operators"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// sandboxLabels are the label selectors used to discover sandbox containers.
var sandboxLabels = map[string]string{
	"dvm.sandbox": "true",
}

// runSandboxGet lists all active sandbox containers by querying the runtime.
func runSandboxGet(cmd *cobra.Command) error {
	ctx := context.Background()

	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		render.Error("Failed to create container runtime")
		render.Plain(FormatSuggestions(SuggestNoContainerRuntime()...))
		return errSilent
	}

	containers, err := runtime.ListContainers(ctx, sandboxLabels)
	if err != nil {
		render.Errorf("Failed to list sandboxes: %v", err)
		return errSilent
	}

	if len(containers) == 0 {
		return render.OutputWith(outputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No sandboxes found",
			EmptyHints:   []string{"dvm sandbox create <language>"},
		})
	}

	// Build table for structured/json/yaml output
	if outputFormat == "json" || outputFormat == "yaml" {
		return renderSandboxStructured(containers)
	}

	return renderSandboxTable(containers)
}

// renderSandboxTable renders sandbox containers as a table.
func renderSandboxTable(containers []operators.ContainerInfo) error {
	headers := []string{"NAME", "LANGUAGE", "VERSION", "STATUS", "IMAGE"}
	var rows [][]string
	for _, c := range containers {
		lang := c.Labels["dvm.sandbox.lang"]
		version := c.Labels["dvm.sandbox.version"]
		rows = append(rows, []string{c.Name, lang, version, c.Status, c.Image})
	}

	return render.OutputWith(outputFormat, render.TableData{
		Headers: headers,
		Rows:    rows,
	}, render.Options{
		Type: render.TypeTable,
	})
}

// renderSandboxStructured renders sandbox containers as JSON or YAML.
func renderSandboxStructured(containers []operators.ContainerInfo) error {
	type sandboxEntry struct {
		Name     string `json:"name" yaml:"name"`
		Language string `json:"language" yaml:"language"`
		Version  string `json:"version" yaml:"version"`
		Status   string `json:"status" yaml:"status"`
		Image    string `json:"image" yaml:"image"`
		ID       string `json:"id" yaml:"id"`
	}

	entries := make([]sandboxEntry, len(containers))
	for i, c := range containers {
		entries[i] = sandboxEntry{
			Name:     c.Name,
			Language: c.Labels["dvm.sandbox.lang"],
			Version:  c.Labels["dvm.sandbox.version"],
			Status:   c.Status,
			Image:    c.Image,
			ID:       c.ID,
		}
	}

	return render.OutputWith(outputFormat, entries, render.Options{})
}
