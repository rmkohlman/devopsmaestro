# opencode CLI Tool

[opencode](https://github.com/sst/opencode) is a terminal AI coding assistant. DevOpsMaestro installs it as an opt-in workspace tool so it is available inside your dev container alongside your code.

---

## Enabling opencode

Add `tools.opencode: true` to your workspace YAML:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-api
spec:
  tools:
    opencode: true
```

Apply and rebuild:

```bash
dvm apply -f workspace.yaml
dvm build
```

The next build installs the opencode binary into the container image. Attach and run it from any terminal:

```bash
dvm attach
# inside the container:
opencode
```

`tools.opencode` defaults to `false`. When it is omitted or set to `false`, the `tools:` section is not emitted in YAML export at all — it never adds bloat to images that do not need it.

---

## What it installs

The Dockerfile builder downloads the [opencode release binary](https://github.com/sst/opencode/releases) for the container's OS and architecture at image build time. No package manager required.

Supported platforms: `linux/amd64`, `linux/arm64`.

---

## API key configuration

opencode reads API keys from environment variables at runtime. Keys are **never baked into the image** — pass them at container start time via dvm credentials.

### Step 1 — Store the key with dvm

```bash
# Anthropic API key (recommended)
dvm create credential ANTHROPIC_API_KEY \
  --source env \
  --env-var ANTHROPIC_API_KEY \
  --workspace dev

# Or OpenAI API key
dvm create credential OPENAI_API_KEY \
  --source env \
  --env-var OPENAI_API_KEY \
  --workspace dev
```

> **Note:** These credentials read from your host environment. Set the key in your shell before running `dvm attach`:
> ```bash
> export ANTHROPIC_API_KEY="sk-ant-..."
> dvm attach
> ```

### Step 2 — Verify inside the container

```bash
echo $ANTHROPIC_API_KEY   # should print the key value
opencode                  # starts the TUI
```

### Supported environment variables

| Variable | Provider |
|----------|----------|
| `ANTHROPIC_API_KEY` | Anthropic (Claude) |
| `OPENAI_API_KEY` | OpenAI (GPT-4) |
| `OPENAI_BASE_URL` | OpenAI-compatible endpoint |
| `GEMINI_API_KEY` | Google Gemini |
| `GROQ_API_KEY` | Groq |
| `OPENROUTER_API_KEY` | OpenRouter |

See the [opencode documentation](https://github.com/sst/opencode) for a full list of supported providers and models.

---

## Using MaestroVault for API keys

If you store secrets in [MaestroVault](https://github.com/rmkohlman/devopsmaestro), use the `vault` source instead:

```bash
dvm create credential ANTHROPIC_API_KEY \
  --source vault \
  --vault-secret "opencode-anthropic-key" \
  --workspace dev
```

The key is fetched from the vault at container start and injected as an environment variable — the value never touches the image.

---

## Corporate proxy considerations

If you are behind a corporate proxy, set standard proxy environment variables in your workspace:

```yaml
spec:
  env:
    HTTP_PROXY: "http://proxy.corp.example.com:8080"
    HTTPS_PROXY: "http://proxy.corp.example.com:8080"
    NO_PROXY: "localhost,127.0.0.1,.corp.example.com"
```

opencode respects `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` for outbound API calls.

---

## Neovim integration

The opencode CLI tool integrates with the `opencode.nvim` Neovim plugin. See [MaestroNvim docs](https://rmkohlman.github.io/MaestroNvim/plugins/opencode/) for setup details.

---

## Workspace YAML example (complete)

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-api
spec:
  tools:
    opencode: true
  build:
    devStage:
      devTools:
        - gopls
        - delve
  shell:
    type: zsh
    framework: oh-my-zsh
  nvim:
    structure: lazyvim
    pluginPackage: rmkohlman
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
    command: ["/bin/zsh", "-l"]
```

---

## Related

- [Credential Reference](../reference/credential.md) — Managing secrets
- [Workspace Reference](../reference/workspace.md) — Full workspace YAML schema
- [Building & Attaching](build-attach.md) — Container lifecycle
