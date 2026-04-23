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
dvm apply -f ~/configs/app.yaml
dvm apply -f ./resources/my-app.yaml
```

Both relative and absolute paths work.

---

## URL Source

Fetch from HTTP/HTTPS URLs:

```bash
dvm apply -f https://raw.githubusercontent.com/user/repo/main/workspace.yaml
dvm apply -f https://example.com/configs/workspace.yaml
```

URLs are auto-detected by the `http://` or `https://` prefix.

---

## GitHub Shorthand

Convenient shorthand for GitHub raw files:

```bash
dvm apply -f github:rmkohlman/devopsmaestro/examples/workspace.yaml
```

This expands to:

```
https://raw.githubusercontent.com/rmkohlman/devopsmaestro/main/examples/workspace.yaml
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
cat workspace.yaml | dvm apply -f -

# Pipe from command
curl -s https://example.com/workspace.yaml | dvm apply -f -

# Here document
dvm apply -f - << 'EOF'
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: my-app
spec:
  domain: myorg
EOF
```

Use `-` as the filename to read from stdin.

---

## Auto-Detection

The source type is automatically detected from the path prefix:

| Prefix | Source Type |
|--------|-------------|
| `-` (literal dash) | Stdin |
| `http://` or `https://` | URL |
| `github:` | GitHub shorthand |
| *(anything else)* | Local file |

---

## Examples

### Apply Resource from GitHub

```bash
dvm apply -f github:rmkohlman/devopsmaestro/examples/workspace.yaml
```

### Apply Multiple from URLs

```bash
dvm apply -f https://example.com/ecosystem.yaml
dvm apply -f https://example.com/domain.yaml
dvm apply -f https://example.com/app.yaml
```

### Scripted Application

```bash
#!/bin/bash
resources=(ecosystem domain app workspace)
for resource in "${resources[@]}"; do
  dvm apply -f "https://example.com/configs/${resource}.yaml"
done
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
