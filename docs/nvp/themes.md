# Built-in Themes

DevOpsMaestro ships with **34+ built-in themes** embedded in the binary. No installation required — every theme is available immediately.

---

## Browsing the Library

```bash
# List all built-in themes
dvm library get themes

# Short alias
dvm lib ls nt

# Show details for a specific theme
dvm library describe theme <name>
```

To see themes alongside any user-defined themes you've created:

```bash
dvm get nvim themes
```

---

## CoolNight Collection (21 Themes)

The CoolNight Collection is a set of parametrically generated themes using calibrated HSL color relationships. Every variant maintains consistent contrast ratios and semantic color mapping, so switching between them feels natural.

**→ [Full CoolNight Documentation](coolnight.md)**

### Blue Family

| Theme | Description | Best For |
|-------|-------------|----------|
| `coolnight-arctic` | Ice-cold crisp blue (190°) | TypeScript, Go, documentation |
| `coolnight-ocean` | Deep professional blue (210°) — **default** | General development, Python |
| `coolnight-midnight` | Dark intense blue (240°) | Late-night coding, C++ |

```bash
dvm set theme coolnight-ocean
dvm set theme coolnight-arctic
dvm set theme coolnight-midnight
```

### Purple Family

| Theme | Description | Best For |
|-------|-------------|----------|
| `coolnight-violet` | Soft violet (270°) | Web development, CSS |
| `coolnight-synthwave` | Neon cyberpunk purple (280°) | JavaScript, creative coding |
| `coolnight-grape` | Rich sophisticated grape (290°) | Rust, systems programming |

```bash
dvm set theme coolnight-synthwave
dvm set theme coolnight-violet
dvm set theme coolnight-grape
```

### Green Family

| Theme | Description | Best For |
|-------|-------------|----------|
| `coolnight-forest` | Earthy forest green (110°) | Bash scripts, DevOps |
| `coolnight-matrix` | High-contrast Matrix green (120°) | Terminal work, security |
| `coolnight-mint` | Fresh modern mint (150°) | React, Vue.js, modern JS |

```bash
dvm set theme coolnight-matrix
dvm set theme coolnight-forest
dvm set theme coolnight-mint
```

### Warm Family

| Theme | Description | Best For |
|-------|-------------|----------|
| `coolnight-ember` | Glowing energetic ember (20°) | Java, Spring Boot |
| `coolnight-sunset` | Warm inviting orange (30°) | HTML, markup languages |
| `coolnight-gold` | Premium golden yellow (45°) | Configuration files, YAML |

```bash
dvm set theme coolnight-sunset
dvm set theme coolnight-ember
dvm set theme coolnight-gold
```

### Red/Pink Family

| Theme | Description | Best For |
|-------|-------------|----------|
| `coolnight-crimson` | Bold deep crimson (0°) | Debugging, error-heavy work |
| `coolnight-sakura` | Elegant cherry blossom (320°) | Design systems |
| `coolnight-rose` | Soft rose pink (350°) | Personal projects |

```bash
dvm set theme coolnight-rose
dvm set theme coolnight-crimson
dvm set theme coolnight-sakura
```

### Monochrome Family

| Theme | Description | Best For |
|-------|-------------|----------|
| `coolnight-mono-charcoal` | Dark charcoal, minimalist | Distraction-free coding |
| `coolnight-mono-slate` | Blue-gray, professional | Enterprise development |
| `coolnight-mono-warm` | Warm gray, comfortable | Long coding sessions |

```bash
dvm set theme coolnight-mono-slate
dvm set theme coolnight-mono-charcoal
dvm set theme coolnight-mono-warm
```

### Special Variants

| Theme | Inspiration | Description |
|-------|-------------|-------------|
| `coolnight-nord` | Nord | Arctic blue-gray Nordic aesthetic |
| `coolnight-dracula` | Dracula | Rich gothic purple |
| `coolnight-solarized` | Solarized | Scientific precision blue-green |

```bash
dvm set theme coolnight-nord
dvm set theme coolnight-dracula
dvm set theme coolnight-solarized
```

---

## Popular Themes (13 others)

In addition to CoolNight, the library includes 13 popular community themes:

| Theme | Plugin | Style | Description |
|-------|--------|-------|-------------|
| `tokyonight-night` | folke/tokyonight.nvim | night | Dark blue, popular choice |
| `tokyonight-ocean` | folke/tokyonight.nvim | moon | Tokyo Night ocean variant |
| `tokyonight-custom` | folke/tokyonight.nvim | — | Custom Tokyo Night variant |
| `catppuccin-mocha` | catppuccin/nvim | mocha | Dark pastel, very popular |
| `catppuccin-latte` | catppuccin/nvim | latte | Light pastel |
| `gruvbox-dark` | ellisonleao/gruvbox.nvim | dark | Classic warm retro dark |
| `nord` | shaunsingh/nord.nvim | — | Arctic blue-gray |
| `dracula` | Mofiqul/dracula.nvim | — | Dark purple gothic |
| `onedark` | navarasu/onedark.nvim | — | Dark blue (Atom-inspired) |
| `rose-pine` | rose-pine/neovim | — | Natural, warm tones |
| `kanagawa` | rebelot/kanagawa.nvim | — | Inspired by Kanagawa art |
| `everforest` | sainnhe/everforest | — | Warm green, eye-friendly |
| `solarized-dark` | ishan9299/nvim-solarized-lua | — | Scientific blue-green dark |

---

## Custom CoolNight Variants

Generate your own CoolNight variant using any hue angle, hex color, or preset name:

```bash
# From a hue angle (0–360)
nvp theme create --from "165" --name coolnight-teal
nvp theme create --from "315" --name coolnight-magenta

# From a hex color
nvp theme create --from "#8B00FF" --name my-violet

# From a preset name
nvp theme create --from "synthwave" --name my-synth

# Create and use immediately
nvp theme create --from "280" --name custom-purple --use
```

---

## Applying Themes

### Via CLI

```bash
# Global default (no flags = global)
dvm set theme coolnight-ocean

# Specific scope
dvm set theme coolnight-synthwave --workspace dev
dvm set theme tokyonight-night --app my-api
dvm set theme gruvbox-dark --domain backend
dvm set theme catppuccin-mocha --ecosystem my-platform
```

### Via YAML

Themes can be embedded directly in resource YAML:

```yaml
# In a Workspace YAML
spec:
  nvim:
    theme: coolnight-synthwave

# In an App YAML
spec:
  theme: tokyonight-night

# In a Domain YAML
spec:
  theme: gruvbox-dark
```

Apply with:

```bash
dvm apply -f workspace.yaml
```

---

## Creating Custom Themes

Define a custom theme as a YAML resource:

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: my-custom-theme
  description: My personalized colorscheme
  category: dark
spec:
  plugin:
    repo: "folke/tokyonight.nvim"
  style: "night"
  colors:
    bg: "#1a1b26"
    fg: "#c0caf5"
    accent: "#7aa2f7"
```

```bash
dvm apply -f my-custom-theme.yaml

# Export an existing theme to customize
dvm get nvim theme coolnight-ocean -o yaml > my-ocean.yaml
# Edit my-ocean.yaml, then:
dvm apply -f my-ocean.yaml
```

User-defined themes with the same name as a built-in theme override the library version.

---

## Related

- **[Quick Start: Themes](quick-start-themes.md)** — Step-by-step guide to changing themes
- **[CoolNight Collection](coolnight.md)** — All 21 CoolNight variants in detail
- **[NvimTheme YAML Reference](../reference/nvim-theme.md)** — Full YAML schema
- **[Theme Hierarchy](../advanced/theme-hierarchy.md)** — How themes cascade through the hierarchy
