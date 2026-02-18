# ColorProvider Integration Guide

This guide demonstrates how to use the new ColorProvider system in DevOpsMaestro CLI commands.

## Overview

The ColorProvider system provides a clean, decoupled way to access theme colors in CLI commands without directly importing theme internals. It supports:

- **Active theme loading** from theme store
- **Specific theme selection** 
- **No-color modes** (`--no-color` flag, `NO_COLOR` env var, `TERM=dumb`)
- **Graceful fallbacks** to default colors
- **Context-based injection** for clean dependency management

## Architecture

```
Command → InitColorProviderForCommand() → Context with ColorProvider
   ↓
ThemeStore → ThemeStoreAdapter → PaletteProvider → ProviderFactory → ColorProvider
```

## Quick Start

### 1. Basic Command Setup

```go
package main

import (
    "context"
    "devopsmaestro/pkg/colors"
)

func myCommand(themePath string, noColor bool) {
    ctx := context.Background()
    
    // Initialize ColorProvider for the command
    ctx, err := colors.InitColorProviderForCommand(ctx, themePath, noColor)
    if err != nil {
        // This should rarely happen - falls back to defaults on errors
        log.Printf("Warning: using default colors: %v", err)
    }
    
    // Use colors in rendering
    renderSomething(ctx)
}

func renderSomething(ctx context.Context) {
    provider := colors.MustFromContext(ctx)
    
    fmt.Printf("%sSuccess!%s\n", provider.Success(), provider.Foreground())
    fmt.Printf("%sError occurred%s\n", provider.Error(), provider.Foreground())
}
```

### 2. Command with Cobra CLI

```go
import (
    "github.com/spf13/cobra"
    "devopsmaestro/pkg/colors"
)

var rootCmd = &cobra.Command{
    Use: "myapp",
    RunE: func(cmd *cobra.Command, args []string) error {
        noColor, _ := cmd.Flags().GetBool("no-color")
        themePath := colors.GetDefaultThemePath() // or get from config
        
        ctx, err := colors.InitColorProviderForCommand(cmd.Context(), themePath, noColor)
        if err != nil {
            return err
        }
        
        return runCommand(ctx, args)
    },
}

func init() {
    rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
}
```

### 3. Specific Theme Usage

```go
func previewTheme(themeName string) error {
    ctx := context.Background()
    themePath := colors.GetDefaultThemePath()
    
    // Load a specific theme (not the active one)
    ctx, err := colors.InitColorProviderWithTheme(ctx, themePath, themeName, false)
    if err != nil {
        return fmt.Errorf("theme %q not found: %w", themeName, err)
    }
    
    provider := colors.MustFromContext(ctx)
    
    // Show theme preview
    fmt.Printf("Theme: %s\n", provider.Name())
    fmt.Printf("Primary: %s●%s\n", provider.Primary(), provider.Foreground())
    fmt.printf("Error: %s●%s\n", provider.Error(), provider.Foreground())
    // ... etc
    
    return nil
}
```

## API Reference

### Helper Functions

#### `InitColorProviderForCommand(ctx, themePath, noColor) (context.Context, error)`
Main initialization function for commands. Never returns an error - falls back to defaults.

**Parameters:**
- `ctx`: Base context to inject ColorProvider into
- `themePath`: Path to theme store (use `GetDefaultThemePath()`)
- `noColor`: Whether to disable colors (`--no-color` flag)

**Returns:**
- Context with ColorProvider injected
- Always nil error (graceful fallback behavior)

#### `InitColorProviderWithTheme(ctx, themePath, themeName, noColor) (context.Context, error)`
Initialize with a specific theme (for preview/testing).

**Parameters:**
- `ctx`: Base context  
- `themePath`: Path to theme store
- `themeName`: Specific theme name to load
- `noColor`: Whether to disable colors

**Returns:**
- Context with ColorProvider injected
- Error if specific theme not found

#### `GetDefaultThemePath() string`
Gets default theme path with proper precedence:
1. `DVM_THEME_PATH` environment variable
2. `$XDG_CONFIG_HOME/dvm/themes`  
3. `$HOME/.config/dvm/themes`
4. `./themes`

#### `IsNoColorRequested(noColorFlag) bool`
Checks if color should be disabled based on:
- `noColorFlag` parameter
- `NO_COLOR` environment variable
- `TERM=dumb` environment variable

### Context Functions

#### `FromContext(ctx) (ColorProvider, bool)`
Safely retrieve ColorProvider from context.

```go
provider, ok := colors.FromContext(ctx)
if !ok {
    // Handle missing provider
}
```

#### `MustFromContext(ctx) ColorProvider`
Retrieve ColorProvider from context, panic if not found.

```go
provider := colors.MustFromContext(ctx) // Use only when provider is guaranteed
```

#### `WithProvider(ctx, provider) context.Context`
Inject ColorProvider into context (mainly for testing).

### ColorProvider Interface

```go
type ColorProvider interface {
    // Primary colors
    Primary() string    // Main brand/accent color
    Secondary() string  // Secondary brand color
    Accent() string     // Highlight/focus color
    
    // Status colors  
    Success() string    // Green-ish success state
    Warning() string    // Yellow-ish warning state
    Error() string      // Red-ish error state
    Info() string       // Blue-ish info state
    
    // UI colors
    Foreground() string // Main text color
    Background() string // Main background color  
    Muted() string      // Subdued/disabled text
    Highlight() string  // Selection/hover background
    Border() string     // Border/separator color
    
    // Metadata
    Name() string       // Theme name
    IsLight() bool      // Whether this is a light theme
}
```

## No-Color Support

The system automatically handles no-color modes:

### Via Command Flag
```bash
myapp --no-color
```

### Via Environment Variable  
```bash
NO_COLOR=1 myapp
```

### Via Terminal Detection
```bash
TERM=dumb myapp  # Automatically disables colors
```

When no-color mode is active, all ColorProvider methods return empty strings, which rendering code should interpret as "no color formatting".

## Example: Status Display

```go
func displayStatus(ctx context.Context, status string, message string) {
    provider := colors.MustFromContext(ctx)
    
    var statusColor string
    switch status {
    case "success":
        statusColor = provider.Success()
    case "error":  
        statusColor = provider.Error()
    case "warning":
        statusColor = provider.Warning()
    case "info":
        statusColor = provider.Info()
    default:
        statusColor = provider.Muted()
    }
    
    reset := provider.Foreground()
    fmt.Printf("[%s%s%s] %s\n", statusColor, status, reset, message)
}

// Usage:
displayStatus(ctx, "success", "Operation completed")
displayStatus(ctx, "error", "Something went wrong")
```

## Testing

### Mock Provider for Tests

```go
func TestMyCommand(t *testing.T) {
    // Use mock provider for predictable colors
    provider := colors.NewMockColorProvider("test-theme", false)
    ctx := colors.WithProvider(context.Background(), provider)
    
    // Test your command logic
    result := myCommandLogic(ctx)
    
    // Assertions can check for specific colors
    assert.Contains(t, result, provider.Success())
}
```

### No-Color Testing

```go  
func TestNoColorMode(t *testing.T) {
    provider := colors.NewNoColorProvider()
    ctx := colors.WithProvider(context.Background(), provider)
    
    result := myCommandLogic(ctx)
    
    // Should contain no ANSI color codes
    assert.NotContains(t, result, "\x1b[")
}
```

## Migration from Direct Theme Usage

### Before (Direct theme import)
```go
import "devopsmaestro/pkg/nvimops/theme" // ❌ Creates coupling

func myCommand() {
    store := theme.NewFileStore(path)
    activeTheme, _ := store.GetActive()
    palette := activeTheme.ToPalette()
    color := palette.Get("primary")
}
```

### After (ColorProvider)  
```go
import "devopsmaestro/pkg/colors" // ✅ Decoupled interface

func myCommand(ctx context.Context) {
    provider := colors.MustFromContext(ctx)
    color := provider.Primary()
}
```

The new approach:
- ✅ **Decouples** commands from theme internals
- ✅ **Supports** no-color modes automatically  
- ✅ **Provides** graceful fallbacks
- ✅ **Enables** better testing with mocks
- ✅ **Maintains** clean dependency boundaries

## Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `DVM_THEME_PATH` | Custom theme store path | `/custom/themes` |
| `NO_COLOR` | Disable color output | `NO_COLOR=1` |
| `XDG_CONFIG_HOME` | XDG config directory | `/home/user/.config` |
| `TERM` | Terminal type | `TERM=dumb` |

## Error Handling

The ColorProvider system is designed for robustness:

- **`InitColorProviderForCommand`** never fails - falls back to defaults
- **`InitColorProviderWithTheme`** fails only for missing themes
- **Missing colors** fall back to reasonable defaults
- **No-color modes** are handled transparently
- **Context missing** panics only with `MustFromContext()`

This ensures commands remain functional even when theme configuration is incomplete or missing.