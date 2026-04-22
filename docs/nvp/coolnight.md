# CoolNight Theme Collection

The CoolNight Collection is a set of 21 parametrically generated themes that provide consistent, professional color schemes optimized for extended development sessions.

---

## Overview

CoolNight themes are designed with:

- **Reduced eye strain** - Carefully calibrated contrast ratios
- **Consistent syntax highlighting** - Uniform color semantics across themes
- **Professional appearance** - Suitable for presentations and screen sharing
- **Parametric generation** - Mathematically derived color relationships
- **Wide variety** - 21 variants covering all color preferences

---

## Theme Philosophy

### Color Science

CoolNight themes use:
- **HSL color space** for predictable hue relationships
- **Consistent lightness** across variants for uniform readability
- **Optimal contrast ratios** meeting WCAG accessibility standards
- **Semantic color mapping** - similar code elements use related colors

### Design Principles

1. **Hierarchy** - Different code elements have clear visual importance
2. **Harmony** - All colors work together aesthetically
3. **Function** - Colors convey meaning (errors=red, strings=green, etc.)
4. **Consistency** - Same color rules across all 21 variants

---

## Complete Collection

### Blue Family (Ocean Tones)

| Theme | Hue | Character | Best For |
|-------|-----|-----------|----------|
| `coolnight-arctic` | 190° | Ice blue, crisp and clean | TypeScript, Go, documentation |
| `coolnight-ocean` | 210° | Deep blue, default variant | General development, Python |
| `coolnight-midnight` | 240° | Dark blue, intense focus | Late-night coding, C++ |

**Preview:**
```bash
nvp theme use coolnight-ocean
nvp theme use coolnight-arctic
nvp theme use coolnight-midnight
```

### Purple Family (Creative Tones)

| Theme | Hue | Character | Best For |
|-------|-----|-----------|----------|
| `coolnight-violet` | 270° | Soft violet, gentle on eyes | Web development, CSS |
| `coolnight-synthwave` | 280° | Neon purple, retro vibes | JavaScript, creative coding |
| `coolnight-grape` | 290° | Rich grape, sophisticated | Rust, systems programming |

**Preview:**
```bash
nvp theme use coolnight-synthwave
nvp theme use coolnight-violet  
nvp theme use coolnight-grape
```

### Green Family (Natural Tones)

| Theme | Hue | Character | Best For |
|-------|-----|-----------|----------|
| `coolnight-forest` | 110° | Forest green, earthy | Bash scripts, DevOps |
| `coolnight-matrix` | 120° | Matrix green, high contrast | Terminal work, cybersec |
| `coolnight-mint` | 150° | Fresh mint, modern | React, Vue.js, modern JS |

**Preview:**
```bash
nvp theme use coolnight-matrix
nvp theme use coolnight-forest
nvp theme use coolnight-mint
```

### Warm Family (Energy Tones)

| Theme | Hue | Character | Best For |
|-------|-----|-----------|----------|
| `coolnight-ember` | 20° | Glowing ember, energetic | Java, Spring Boot |
| `coolnight-sunset` | 30° | Warm orange, inviting | HTML, markup languages |
| `coolnight-gold` | 45° | Golden yellow, premium | Configuration files, YAML |

**Preview:**
```bash
nvp theme use coolnight-sunset
nvp theme use coolnight-ember
nvp theme use coolnight-gold
```

### Red/Pink Family (Passionate Tones)

| Theme | Hue | Character | Best For |
|-------|-----|-----------|----------|
| `coolnight-crimson` | 0° | Deep crimson, bold | Error handling, debugging |
| `coolnight-sakura` | 320° | Cherry blossom, elegant | Design systems, Figma |
| `coolnight-rose` | 350° | Rose pink, romantic | Personal projects, blogs |

**Preview:**
```bash
nvp theme use coolnight-rose
nvp theme use coolnight-crimson
nvp theme use coolnight-sakura
```

### Monochrome Family (Focus Tones)

| Theme | Character | Best For |
|-------|-----------|----------|
| `coolnight-mono-charcoal` | Charcoal gray, minimalist | Distraction-free coding |
| `coolnight-mono-slate` | Slate gray, professional | Enterprise development |
| `coolnight-mono-warm` | Warm gray, comfortable | Long coding sessions |

**Preview:**
```bash
nvp theme use coolnight-mono-slate
nvp theme use coolnight-mono-charcoal
nvp theme use coolnight-mono-warm
```

### Special Variants (Inspired Themes)

| Theme | Inspiration | Character | Best For |
|-------|-------------|-----------|----------|
| `coolnight-nord` | Nord theme | Arctic blue-gray | Clean, Nordic aesthetic |
| `coolnight-dracula` | Dracula theme | Rich purple | Dark, gothic feel |
| `coolnight-solarized` | Solarized theme | Scientific precision | Academic, research |

**Preview:**
```bash
nvp theme use coolnight-nord
nvp theme use coolnight-dracula  
nvp theme use coolnight-solarized
```

---

## Parametric Generator

### Create Custom Variants

Generate your own CoolNight variant using a hue angle (0–360), a hex color, or a preset name:

```bash
# Create by hue angle
nvp theme create --from "75" --name coolnight-lime     # Lime green
nvp theme create --from "165" --name coolnight-teal    # Teal blue
nvp theme create --from "315" --name coolnight-magenta # Hot magenta

# Create from a hex color
nvp theme create --from "#8B00FF" --name coolnight-violet

# Create from a preset name
nvp theme create --from "synthwave" --name my-synth

# Use immediately after creation
nvp theme use coolnight-lime
```

### Hue Reference

| Hue Range | Color Family | Examples |
|-----------|--------------|----------|
| 0° - 30° | Red to Orange | crimson, ember, sunset |
| 30° - 90° | Orange to Yellow | gold, warm yellows |
| 90° - 150° | Yellow to Green | forest, matrix, mint |
| 150° - 210° | Green to Blue | teal, arctic, ocean |
| 210° - 270° | Blue to Purple | midnight, violet |
| 270° - 330° | Purple to Pink | synthwave, grape, sakura |
| 330° - 360° | Pink to Red | rose, back to crimson |

### Advanced Generator Options

The `--from` flag accepts a hue angle, hex color, or preset name. For full control, create a theme YAML and apply it:

```bash
# Use a specific hue angle
nvp theme create --from "210" --name my-ocean

# Use a hex color as the base
nvp theme create --from "#0a7fa8" --name my-teal

# Use a built-in preset name
nvp theme create --from "ocean" --name ocean-custom

# Activate immediately after creation
nvp theme create --from "280" --name custom-purple --use
```

---

## Color Palette Structure

### Semantic Color Mapping

Every CoolNight theme uses this consistent mapping:

| Semantic Role | Purpose | Example Elements |
|---------------|---------|------------------|
| `bg` | Background | Editor background, panels |
| `fg` | Foreground | Default text, variables |
| `accent` | Primary accent | Cursor, selection, highlights |
| `comment` | Comments | `// comments`, `# comments` |
| `keyword` | Language keywords | `function`, `class`, `if`, `while` |
| `string` | String literals | `"hello"`, `'world'` |
| `function` | Function names | `myFunction()`, method calls |
| `type` | Type annotations | `String`, `int`, class names |
| `constant` | Constants | `true`, `false`, `null`, numbers |
| `error` | Error indicators | Error squiggles, diagnostics |
| `warning` | Warning indicators | Warning messages |
| `info` | Information | Hints, info messages |
| `selection` | Text selection | Selected text background |
| `border` | UI borders | Window borders, splits |

### Color Relationships

The parametric generator maintains these relationships:

- **Accent color** derived from primary hue
- **Syntax colors** are hue variations (±30°, ±60°, etc.)
- **UI colors** use desaturated versions of the primary hue
- **Semantic colors** (error, warning) use appropriate hues regardless of theme

---

## Usage Recommendations

### By Development Environment

**Terminal-heavy workflows:**
- `coolnight-matrix` - High contrast green
- `coolnight-mono-charcoal` - Minimal distractions

**Web development:**
- `coolnight-mint` - Modern, fresh feel
- `coolnight-synthwave` - Creative, vibrant

**Systems programming:**
- `coolnight-midnight` - Deep focus
- `coolnight-grape` - Sophisticated, serious

**Documentation writing:**
- `coolnight-arctic` - Clean, readable
- `coolnight-mono-warm` - Easy on eyes

**Presentations/screen sharing:**
- `coolnight-ocean` - Professional default
- `coolnight-sunset` - Warm, welcoming

### By Time of Day

**Morning coding:**
- `coolnight-arctic` - Fresh, energizing
- `coolnight-mint` - Bright start

**Daytime work:**
- `coolnight-ocean` - Balanced, professional
- `coolnight-forest` - Natural, comfortable

**Evening sessions:**
- `coolnight-sunset` - Warm transition
- `coolnight-ember` - Cozy coding

**Late-night coding:**
- `coolnight-midnight` - Deep focus
- `coolnight-mono-slate` - Reduced stimulation

---

## Integration Examples

### With Theme Hierarchy

```bash
# Set different CoolNight variants by context
dvm set theme coolnight-ocean --ecosystem corporate      # Professional default
dvm set theme coolnight-matrix --domain security        # High contrast for security work
dvm set theme coolnight-synthwave --app creative-tool   # Creative project gets creative theme
```

### With Development Workflow

```bash
# Different themes for different branches
git checkout main && dvm set theme coolnight-ocean
git checkout feature/ui && dvm set theme coolnight-mint  
git checkout hotfix/critical && dvm set theme coolnight-crimson
```

### Export for Team

```bash
# Export your favorite CoolNight variant
dvm get nvim theme coolnight-synthwave -o yaml > team-theme.yaml

# Team members apply it
dvm apply -f team-theme.yaml
```

---

## Technical Details

### Color Space

CoolNight uses **HSL (Hue, Saturation, Lightness)** color space:
- **Hue**: 0-360° (color wheel position)  
- **Saturation**: 40-70% (balanced vibrancy)
- **Lightness**: 45-85% (optimal contrast)

### Accessibility

All CoolNight themes meet:
- **WCAG AA** contrast ratios (4.5:1 minimum)
- **Colorblind friendly** - not relying solely on color
- **Reduced motion** - subtle animations only

### Performance

- **CSS custom properties** for easy browser integration
- **Terminal color mapping** for consistent terminal themes
- **Fast switching** - themes cached for instant preview

---

## Troubleshooting

### Theme Not Applying

```bash
# Check if theme exists
nvp theme get | grep coolnight-ocean

# Verify theme content
nvp theme get coolnight-ocean -o yaml

# Regenerate configuration
nvp generate
```

### Colors Look Wrong

```bash
# Check terminal color support
echo $COLORTERM  # Should show: truecolor

# Check Neovim termguicolors
nvim -c 'set termguicolors?' -c 'q'

# Preview the theme to verify colors in terminal
nvp theme preview coolnight-ocean
```

### Custom Variant Issues

```bash
# Check what was generated
nvp theme get my-custom-theme -o yaml

# Regenerate Lua files after making changes
nvp generate
```

---

## Next Steps

- [Theme Hierarchy](../advanced/theme-hierarchy.md) - Cascade themes through your organization
- [Theme IaC](../advanced/theme-iac.md) - Infrastructure as Code for themes
- [All Themes](themes.md) - Complete theme documentation
- [WezTerm Integration](../configuration/wezterm.md) - Terminal theme integration