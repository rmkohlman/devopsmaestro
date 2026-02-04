# Source Types

How DevOpsMaestro resolves input sources for the `-f` flag.

---

## Overview

The `-f` flag accepts multiple source types, auto-detected from the path:

| Source Type | Example | Description |
|-------------|---------|-------------|
| **File** | `-f plugin.yaml` | Local file path |
| **URL** | `-f https://example.com/plugin.yaml` | HTTP/HTTPS URL |
| **GitHub** | `-f github:user/repo/path.yaml` | GitHub shorthand |
| **Stdin** | `-f -` | Read from stdin |

---

## File Source

Read from local filesystem:

```bash
dvm apply -f workspace.yaml
nvp apply -f ~/configs/telescope.yaml
nvp apply -f ./plugins/my-plugin.yaml
```

Both relative and absolute paths work.

---

## URL Source

Fetch from HTTP/HTTPS URLs:

```bash
nvp apply -f https://raw.githubusercontent.com/user/repo/main/plugin.yaml
nvp apply -f https://example.com/configs/workspace.yaml
```

URLs are auto-detected by the `http://` or `https://` prefix.

---

## GitHub Shorthand

Convenient shorthand for GitHub raw files:

```bash
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
```

This expands to:

```
https://raw.githubusercontent.com/rmkohlman/nvim-yaml-plugins/main/plugins/telescope.yaml
```

### Format

```
github:<owner>/<repo>/<path>
```

- Assumes `main` branch
- Converts to raw.githubusercontent.com URL

---

## Stdin Source

Read from standard input:

```bash
# Pipe from file
cat plugin.yaml | nvp apply -f -

# Pipe from command
curl -s https://example.com/plugin.yaml | nvp apply -f -

# Here document
nvp apply -f - << 'EOF'
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: my-plugin
spec:
  repo: user/my-plugin
EOF
```

Use `-` as the filename to read from stdin.

---

## Auto-Detection

The source type is automatically detected:

```go
func DetectSourceType(path string) SourceType {
    if path == "-" {
        return SourceTypeStdin
    }
    if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
        return SourceTypeURL
    }
    if strings.HasPrefix(path, "github:") {
        return SourceTypeGitHub
    }
    return SourceTypeFile
}
```

---

## Examples

### Install Plugin from GitHub

```bash
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
```

### Apply Multiple from URLs

```bash
nvp apply -f https://example.com/telescope.yaml
nvp apply -f https://example.com/treesitter.yaml
nvp apply -f https://example.com/lspconfig.yaml
```

### Scripted Installation

```bash
#!/bin/bash
plugins=(telescope treesitter lspconfig nvim-cmp)
for plugin in "${plugins[@]}"; do
  nvp apply -f "github:rmkohlman/nvim-yaml-plugins/plugins/${plugin}.yaml"
done
nvp generate
```

### From Curl

```bash
curl -s https://example.com/my-config.yaml | dvm apply -f -
```

---

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| `file not found` | File doesn't exist | Check path |
| `failed to fetch URL` | Network error or 404 | Check URL, network |
| `invalid YAML` | Malformed YAML | Validate YAML syntax |
| `unknown kind` | Unsupported resource type | Check `kind` field |

---

## Next Steps

- [Architecture](architecture.md) - Internal architecture
- [dvm Commands](../dvm/commands.md) - Command reference
- [nvp Commands](../nvp/commands.md) - Command reference
