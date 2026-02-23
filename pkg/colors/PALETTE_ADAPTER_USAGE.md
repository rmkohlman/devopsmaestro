# ColorToPaletteAdapter Usage Examples

The `ColorToPaletteAdapter` provides a clean way to convert any `ColorProvider` into a `*palette.Palette` for use with terminal rendering components like `StarshipRenderer`.

## Basic Usage

### Using the Adapter Pattern

```go
import (
    "context"
    "devopsmaestro/pkg/colors"
    "devopsmaestro/pkg/starship"
)

func RenderPrompt(ctx context.Context) error {
    // Get ColorProvider from context
    provider := colors.MustFromContext(ctx)
    
    // Create adapter to convert ColorProvider to palette
    adapter := colors.NewColorToPaletteAdapter(provider)
    palette := adapter.ToPalette()
    
    // Use palette with StarshipRenderer
    renderer := starship.NewRenderer(palette)
    return renderer.Generate("~/.config/starship.toml")
}
```

### Using the Convenience Function

For simple cases, use the `ToPalette()` convenience function:

```go
func QuickPrompt(ctx context.Context) error {
    provider := colors.FromContext(ctx)
    
    // Convert in one call
    palette := colors.ToPalette(provider)
    
    renderer := starship.NewRenderer(palette)
    return renderer.Generate("~/.config/starship.toml")
}
```

## Color Mappings

The adapter maps `ColorProvider` methods to `palette.Palette` semantic constants:

| ColorProvider Method | Palette Constant | Purpose |
|---------------------|------------------|---------|
| `Primary()`         | `palette.ColorPrimary` | Main accent/brand color |
| `Secondary()`       | `palette.ColorSecondary` | Secondary accent |
| `Accent()`          | `palette.ColorAccent` | Highlight/focus color |
| `Success()`         | `palette.ColorSuccess` | Success state (green) |
| `Warning()`         | `palette.ColorWarning` | Warning state (yellow) |
| `Error()`           | `palette.ColorError` | Error state (red) |
| `Info()`            | `palette.ColorInfo` | Info state (blue) |
| `Foreground()`      | `palette.ColorFg` | Main text color |
| `Background()`      | `palette.ColorBg` | Main background |
| `Muted()`           | `palette.ColorComment` | Subdued/disabled text |
| `Highlight()`       | `palette.ColorBgHighlight` | Selection background |
| `Border()`          | `palette.ColorBorder` | Border/separator color |

## Handling NoColorProvider

The adapter gracefully handles `NoColorProvider` (when `--no-color` is set):

```go
// When colors are disabled
provider := colors.NewNoColorProvider()
palette := colors.ToPalette(provider)

// palette.Colors will be empty (not nil)
// This allows rendering to continue without colors
fmt.Printf("Colors available: %d\n", len(palette.Colors)) // Output: 0
```

## Testing with MockColorProvider

```go
func TestMyRenderer(t *testing.T) {
    // Create mock with specific colors
    provider := colors.NewMockColorProvider(
        colors.WithMockName("test-theme"),
        colors.WithMockColor("primary", "#ff0000"),
        colors.WithMockColor("success", "#00ff00"),
    )
    
    palette := colors.ToPalette(provider)
    
    // Test rendering with known colors
    assert.Equal(t, "#ff0000", palette.Colors[palette.ColorPrimary])
    assert.Equal(t, "#00ff00", palette.Colors[palette.ColorSuccess])
}
```

## Integration with Terminal Rendering

### StarshipRenderer Example

```go
type StarshipRenderer struct {
    palette *palette.Palette
}

func NewStarshipRenderer(ctx context.Context) *StarshipRenderer {
    provider := colors.FromContext(ctx)
    return &StarshipRenderer{
        palette: colors.ToPalette(provider),
    }
}

func (s *StarshipRenderer) GeneratePromptSection(status string) string {
    var color string
    switch status {
    case "success":
        color = s.palette.GetOrDefault(palette.ColorSuccess, "#00ff00")
    case "error":
        color = s.palette.GetOrDefault(palette.ColorError, "#ff0000")
    case "warning":
        color = s.palette.GetOrDefault(palette.ColorWarning, "#ffff00")
    default:
        color = s.palette.GetOrDefault(palette.ColorFg, "#ffffff")
    }
    
    return fmt.Sprintf("[%s](%s)", status, color)
}
```

## Architecture Benefits

1. **Separation of Concerns**: ColorProvider interface stays clean and focused
2. **Adapter Pattern**: Bridges different interface requirements without modification
3. **No Color Handling**: Gracefully handles disabled colors
4. **Default Fallbacks**: Uses palette's `GetOrDefault()` for missing colors
5. **Testing**: Easy to test with MockColorProvider

## Performance Notes

- The adapter creates a new palette each time `ToPalette()` is called
- For high-frequency usage, consider caching the palette
- Empty colors are filtered out to reduce memory usage
- Color strings are not validated (assumes valid hex colors from providers)