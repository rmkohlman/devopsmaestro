package cmd

import (
	"fmt"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/render"
)

// tableBuilder is the interface every resource-specific table builder must satisfy.
// Headers returns the column headers for the table (wide adds extra columns).
// Row converts any model value into a slice of string cells.
type tableBuilder interface {
	Headers(wide bool) []string
	Row(model any, wide bool) []string
}

// BuildTable assembles a render.TableData from a builder and a typed slice of items.
// Each item is passed as-is to builder.Row, so builders must type-assert internally.
func BuildTable[T any](builder tableBuilder, items []T, wide bool) render.TableData {
	td := render.TableData{
		Headers: builder.Headers(wide),
		Rows:    make([][]string, 0, len(items)),
	}
	for _, item := range items {
		td.Rows = append(td.Rows, builder.Row(item, wide))
	}
	return td
}

// renderTable writes a render.TableData to stdout using the current output format.
func renderTable(tableData render.TableData) error {
	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

// =============================================================================
// Shared helpers
// =============================================================================

// truncateRight truncates s so that the total display length is maxLen by
// keeping the last (maxLen-3) chars and prefixing with "...".
// Truncation is triggered when len(s) > (maxLen-3), i.e., when the string
// is too long to fit without the prefix. If len(s) <= (maxLen-3) the original
// string is returned unchanged.
func truncateRight(s string, maxLen int) string {
	keep := maxLen - 3
	if len(s) <= keep {
		return s
	}
	return "..." + s[len(s)-keep:]
}

// truncateLeft truncates s to maxLen by keeping the first (maxLen-3) chars
// and suffixing with "...". If len(s) <= maxLen the original string is returned.
func truncateLeft(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// activeMarker returns "● " + name when the IDs match, otherwise just name.
func activeMarker(name string, itemID int, activeID *int) string {
	if activeID != nil && *activeID == itemID {
		return "● " + name
	}
	return name
}

// activeMarkerByName returns "● " + name when names match, otherwise just name.
func activeMarkerByName(name string, activeName string) string {
	if activeName != "" && activeName == name {
		return "● " + name
	}
	return name
}

// splitStatusUptime splits a "state/uptime" status map value.
// Returns ("", "") if the value is empty or has no "/" separator.
func splitStatusUptime(val string) (state, uptime string) {
	if val == "" {
		return "", ""
	}
	parts := strings.SplitN(val, "/", 2)
	state = parts[0]
	if len(parts) == 2 {
		uptime = parts[1]
	}
	return
}

// =============================================================================
// ecosystemTableBuilder
// =============================================================================

// ecosystemTableBuilder builds table rows for *models.Ecosystem values.
// ActiveID, when non-nil, marks the matching ecosystem with a "●" prefix.
type ecosystemTableBuilder struct {
	ActiveID *int
}

func (b *ecosystemTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "DESCRIPTION", "CREATED"}
	if wide {
		headers = append(headers, "ID")
	}
	return headers
}

func (b *ecosystemTableBuilder) Row(model any, wide bool) []string {
	eco := model.(*models.Ecosystem)

	name := activeMarker(eco.Name, eco.ID, b.ActiveID)

	desc := ""
	if eco.Description.Valid {
		desc = truncateLeft(eco.Description.String, 40)
	}

	created := eco.CreatedAt.Format("2006-01-02 15:04")

	row := []string{name, desc, created}
	if wide {
		row = append(row, fmt.Sprintf("%d", eco.ID))
	}
	return row
}

// =============================================================================
// domainTableBuilder
// =============================================================================

// domainTableBuilder builds table rows for *models.Domain values.
type domainTableBuilder struct {
	DataStore db.DataStore
	ActiveID  *int
}

func (b *domainTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "ECOSYSTEM", "DESCRIPTION", "CREATED"}
	if wide {
		headers = append(headers, "ID")
	}
	return headers
}

func (b *domainTableBuilder) Row(model any, wide bool) []string {
	domain := model.(*models.Domain)

	name := activeMarker(domain.Name, domain.ID, b.ActiveID)

	ecoName := ""
	if b.DataStore != nil {
		if eco, err := b.DataStore.GetEcosystemByID(domain.EcosystemID); err == nil {
			ecoName = eco.Name
		}
	}

	desc := ""
	if domain.Description.Valid {
		desc = truncateLeft(domain.Description.String, 30)
	}

	created := domain.CreatedAt.Format("2006-01-02 15:04")

	row := []string{name, ecoName, desc, created}
	if wide {
		row = append(row, fmt.Sprintf("%d", domain.ID))
	}
	return row
}

// =============================================================================
// appTableBuilder
// =============================================================================

// appTableBuilder builds table rows for *models.App values.
type appTableBuilder struct {
	DataStore db.DataStore
	ActiveID  *int
}

func (b *appTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "DOMAIN", "PATH", "CREATED"}
	if wide {
		headers = append(headers, "ID", "GITREPO")
	}
	return headers
}

func (b *appTableBuilder) Row(model any, wide bool) []string {
	app := model.(*models.App)

	name := activeMarker(app.Name, app.ID, b.ActiveID)

	domainName := ""
	if b.DataStore != nil {
		if domain, err := b.DataStore.GetDomainByID(app.DomainID); err == nil {
			domainName = domain.Name
		}
	}

	path := truncateRight(app.Path, 40)
	created := app.CreatedAt.Format("2006-01-02 15:04")

	row := []string{name, domainName, path, created}
	if wide {
		gitRepo := "<none>"
		if app.GitRepoID.Valid {
			gitRepo = fmt.Sprintf("%d", app.GitRepoID.Int64)
		}
		row = append(row, fmt.Sprintf("%d", app.ID), gitRepo)
	}
	return row
}

// =============================================================================
// workspaceTableBuilder
// =============================================================================

// workspaceTableBuilder builds table rows for *models.Workspace values.
// Active workspace is matched by name, not ID.
type workspaceTableBuilder struct {
	DataStore           db.DataStore
	ActiveWorkspaceName string
}

func (b *workspaceTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "APP", "IMAGE", "STATUS"}
	if wide {
		headers = append(headers, "CREATED", "CONTAINER-ID")
	}
	return headers
}

func (b *workspaceTableBuilder) Row(model any, wide bool) []string {
	ws := model.(*models.Workspace)

	name := activeMarkerByName(ws.Name, b.ActiveWorkspaceName)

	appName := ""
	if b.DataStore != nil {
		if app, err := b.DataStore.GetAppByID(ws.AppID); err == nil {
			appName = app.Name
		}
	}

	row := []string{name, appName, ws.ImageName, ws.Status}
	if wide {
		created := ws.CreatedAt.Format("2006-01-02 15:04")
		containerID := "<none>"
		if ws.ContainerID.Valid && ws.ContainerID.String != "" {
			cid := ws.ContainerID.String
			if len(cid) > 12 {
				cid = cid[:12]
			}
			containerID = cid
		}
		row = append(row, created, containerID)
	}
	return row
}

// =============================================================================
// credentialTableBuilder
// =============================================================================

// credentialTableBuilder builds table rows for *models.CredentialDB values.
// Uses the package-level resolveScopeName and formatTargetVars helpers.
type credentialTableBuilder struct {
	DataStore db.DataStore
}

func (b *credentialTableBuilder) Headers(wide bool) []string {
	return []string{"NAME", "SCOPE", "SOURCE", "TARGET", "DESCRIPTION"}
}

func (b *credentialTableBuilder) Row(model any, wide bool) []string {
	cred := model.(*models.CredentialDB)

	scope := ""
	if b.DataStore != nil {
		scope = resolveScopeName(b.DataStore, cred.ScopeType, cred.ScopeID)
	}

	target := formatTargetVars(cred)

	desc := ""
	if cred.Description != nil {
		desc = *cred.Description
	}

	return []string{cred.Name, scope, cred.Source, target, desc}
}

// =============================================================================
// registryTableBuilder
// =============================================================================

// registryTableBuilder builds table rows for *models.Registry values.
// StatusMap pre-resolves live state/uptime strings keyed by registry name.
type registryTableBuilder struct {
	StatusMap map[string]string
}

func (b *registryTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME"}
	if wide {
		headers = append(headers, "CREATED")
	}
	return headers
}

func (b *registryTableBuilder) Row(model any, wide bool) []string {
	reg := model.(*models.Registry)

	state := reg.Status
	uptime := ""
	if b.StatusMap != nil {
		if val, ok := b.StatusMap[reg.Name]; ok {
			s, u := splitStatusUptime(val)
			state = s
			uptime = u
		}
	}

	row := []string{
		reg.Name,
		reg.Type,
		reg.Version,
		fmt.Sprintf("%d", reg.Port),
		reg.Lifecycle,
		state,
		uptime,
	}
	if wide {
		row = append(row, reg.CreatedAt)
	}
	return row
}

// =============================================================================
// gitRepoTableBuilder
// =============================================================================

// gitRepoTableBuilder builds table rows for *models.GitRepoDB values.
type gitRepoTableBuilder struct{}

func (b *gitRepoTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "URL", "STATUS", "LAST_SYNCED"}
	if wide {
		headers = append(headers, "SLUG", "REF", "AUTO_SYNC")
	}
	return headers
}

func (b *gitRepoTableBuilder) Row(model any, wide bool) []string {
	repo := model.(*models.GitRepoDB)

	lastSynced := "never"
	if repo.LastSyncedAt.Valid {
		lastSynced = repo.LastSyncedAt.Time.Format("2006-01-02 15:04")
	}

	row := []string{repo.Name, repo.URL, repo.SyncStatus, lastSynced}
	if wide {
		autoSync := "no"
		if repo.AutoSync {
			autoSync = "yes"
		}
		row = append(row, repo.Slug, repo.DefaultRef, autoSync)
	}
	return row
}

// =============================================================================
// nvimPluginTableBuilder
// =============================================================================

// nvimPluginTableBuilder builds table rows for *models.NvimPluginDB values.
// No wide mode distinction (same columns either way).
type nvimPluginTableBuilder struct{}

func (b *nvimPluginTableBuilder) Headers(wide bool) []string {
	return []string{"NAME", "CATEGORY", "REPO", "VERSION"}
}

func (b *nvimPluginTableBuilder) Row(model any, wide bool) []string {
	plugin := model.(*models.NvimPluginDB)

	mark := " ✗"
	if plugin.Enabled {
		mark = " ✓"
	}
	name := plugin.Name + mark

	category := ""
	if plugin.Category.Valid {
		category = plugin.Category.String
	}

	version := "latest"
	if plugin.Version.Valid && plugin.Version.String != "" {
		version = plugin.Version.String
	} else if plugin.Branch.Valid && plugin.Branch.String != "" {
		version = "branch:" + plugin.Branch.String
	}

	return []string{name, category, plugin.Repo, version}
}

// =============================================================================
// nvimThemeTableBuilder
// =============================================================================

// nvimThemeTableBuilder builds table rows for *models.NvimThemeDB values.
// No wide mode distinction (same columns either way).
type nvimThemeTableBuilder struct{}

func (b *nvimThemeTableBuilder) Headers(wide bool) []string {
	return []string{"NAME", "CATEGORY", "PLUGIN", "STYLE"}
}

func (b *nvimThemeTableBuilder) Row(model any, wide bool) []string {
	theme := model.(*models.NvimThemeDB)

	category := "-"
	if theme.Category.Valid && theme.Category.String != "" {
		category = theme.Category.String
	}

	style := "default"
	if theme.Style.Valid && theme.Style.String != "" {
		style = theme.Style.String
	}

	return []string{theme.Name, category, theme.PluginRepo, style}
}

// =============================================================================
// nvimPackageTableBuilder
// =============================================================================

// nvimPackageTableBuilder builds table rows for *models.NvimPackageDB values.
type nvimPackageTableBuilder struct{}

func (b *nvimPackageTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "CATEGORY", "PLUGINS", "EXTENDS"}
	if wide {
		headers = append(headers, "LABELS")
	}
	return headers
}

func (b *nvimPackageTableBuilder) Row(model any, wide bool) []string {
	pkg := model.(*models.NvimPackageDB)

	category := "-"
	if pkg.Category.Valid && pkg.Category.String != "" {
		category = pkg.Category.String
	}

	pluginCount := fmt.Sprintf("%d", len(pkg.GetPlugins()))

	extends := "-"
	if pkg.Extends.Valid && pkg.Extends.String != "" {
		extends = pkg.Extends.String
	}

	row := []string{pkg.Name, category, pluginCount, extends}
	if wide {
		labels := "-"
		l := pkg.GetLabels()
		if len(l) > 0 {
			parts := make([]string, 0, len(l))
			for k, v := range l {
				parts = append(parts, k+"="+v)
			}
			labels = strings.Join(parts, ",")
		}
		row = append(row, labels)
	}
	return row
}

// =============================================================================
// terminalPromptTableBuilder
// =============================================================================

// terminalPromptTableBuilder builds table rows for *models.TerminalPromptDB values.
type terminalPromptTableBuilder struct{}

func (b *terminalPromptTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "TYPE", "CATEGORY", "ENABLED"}
	if wide {
		headers = append(headers, "PALETTE_REF")
	}
	return headers
}

func (b *terminalPromptTableBuilder) Row(model any, wide bool) []string {
	prompt := model.(*models.TerminalPromptDB)

	category := "-"
	if prompt.Category.Valid && prompt.Category.String != "" {
		category = prompt.Category.String
	}

	enabled := "✗"
	if prompt.Enabled {
		enabled = "✓"
	}

	row := []string{prompt.Name, prompt.Type, category, enabled}
	if wide {
		paletteRef := "-"
		if prompt.PaletteRef.Valid && prompt.PaletteRef.String != "" {
			paletteRef = prompt.PaletteRef.String
		}
		row = append(row, paletteRef)
	}
	return row
}

// =============================================================================
// terminalPackageTableBuilder
// =============================================================================

// terminalPackageTableBuilder builds table rows for *models.TerminalPackageDB values.
type terminalPackageTableBuilder struct{}

func (b *terminalPackageTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "CATEGORY", "PLUGINS", "PROMPTS", "EXTENDS"}
	if wide {
		headers = append(headers, "PROFILES", "LABELS")
	}
	return headers
}

func (b *terminalPackageTableBuilder) Row(model any, wide bool) []string {
	pkg := model.(*models.TerminalPackageDB)

	category := "-"
	if pkg.Category.Valid && pkg.Category.String != "" {
		category = pkg.Category.String
	}

	pluginCount := fmt.Sprintf("%d", len(pkg.GetPlugins()))
	promptCount := fmt.Sprintf("%d", len(pkg.GetPrompts()))

	extends := "-"
	if pkg.Extends.Valid && pkg.Extends.String != "" {
		extends = pkg.Extends.String
	}

	row := []string{pkg.Name, category, pluginCount, promptCount, extends}
	if wide {
		profileCount := fmt.Sprintf("%d", len(pkg.GetProfiles()))
		labels := "-"
		l := pkg.GetLabels()
		if len(l) > 0 {
			parts := make([]string, 0, len(l))
			for k, v := range l {
				parts = append(parts, k+"="+v)
			}
			labels = strings.Join(parts, ",")
		}
		row = append(row, profileCount, labels)
	}
	return row
}

// =============================================================================
// terminalPluginTableBuilder
// =============================================================================

// terminalPluginTableBuilder builds table rows for *models.TerminalPluginDB values.
type terminalPluginTableBuilder struct{}

func (b *terminalPluginTableBuilder) Headers(wide bool) []string {
	headers := []string{"NAME", "REPO", "SHELL", "MANAGER", "ENABLED"}
	if wide {
		headers = append(headers, "CATEGORY", "DESCRIPTION")
	}
	return headers
}

func (b *terminalPluginTableBuilder) Row(model any, wide bool) []string {
	plugin := model.(*models.TerminalPluginDB)

	enabled := "✗"
	if plugin.Enabled {
		enabled = "✓"
	}

	row := []string{plugin.Name, plugin.Repo, plugin.Shell, plugin.Manager, enabled}
	if wide {
		category := "-"
		if plugin.Category.Valid && plugin.Category.String != "" {
			category = plugin.Category.String
		}
		description := "-"
		if plugin.Description.Valid && plugin.Description.String != "" {
			description = plugin.Description.String
		}
		row = append(row, category, description)
	}
	return row
}

// =============================================================================
// caCertTableBuilder
// =============================================================================

// scopedCACert holds a CA cert with its scope label for display in get all.
type scopedCACert struct {
	Name  string
	Scope string
}

// caCertTableBuilder builds table rows for scopedCACert values.
type caCertTableBuilder struct{}

func (b *caCertTableBuilder) Headers(wide bool) []string {
	return []string{"NAME", "SCOPE"}
}

func (b *caCertTableBuilder) Row(model any, wide bool) []string {
	cert := model.(scopedCACert)
	return []string{cert.Name, cert.Scope}
}

// =============================================================================
// buildArgTableBuilder
// =============================================================================

// scopedBuildArg holds a build arg with its scope label for display in get all.
type scopedBuildArg struct {
	Key   string
	Scope string
}

// buildArgTableBuilder builds table rows for scopedBuildArg values.
type buildArgTableBuilder struct{}

func (b *buildArgTableBuilder) Headers(wide bool) []string {
	return []string{"KEY", "SCOPE"}
}

func (b *buildArgTableBuilder) Row(model any, wide bool) []string {
	arg := model.(scopedBuildArg)
	return []string{arg.Key, arg.Scope}
}

// =============================================================================
// crdTableBuilder
// =============================================================================

// crdTableBuilder builds table rows for *models.CustomResourceDefinition values.
type crdTableBuilder struct{}

func (b *crdTableBuilder) Headers(wide bool) []string {
	headers := []string{"KIND", "GROUP", "SCOPE", "PLURAL"}
	if wide {
		headers = append(headers, "SINGULAR", "CREATED")
	}
	return headers
}

func (b *crdTableBuilder) Row(model any, wide bool) []string {
	crd := model.(*models.CustomResourceDefinition)

	row := []string{crd.Kind, crd.Group, crd.Scope, crd.Plural}
	if wide {
		row = append(row, crd.Singular, crd.CreatedAt.Format("2006-01-02 15:04"))
	}
	return row
}
