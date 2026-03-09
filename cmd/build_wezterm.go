package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/pkg/terminalops/wezterm"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// generateWezTermConfig creates a WezTerm configuration file if a terminal emulator config exists in database
func generateWezTermConfig(stagingDir, appName, workspaceName string, ds db.DataStore) error {
	// 1. Look for workspace-specific emulator first
	//    Pattern: "{app}-{workspace}" or "{workspace}"
	workspaceEmulatorName := fmt.Sprintf("%s-%s", appName, workspaceName)
	emulatorDB, err := ds.GetTerminalEmulator(workspaceEmulatorName)
	if err != nil {
		// Try just workspace name
		emulatorDB, err = ds.GetTerminalEmulator(workspaceName)
		if err != nil {
			// 2. Fall back to default emulator if set
			defaultEmulatorName, err := ds.GetDefault("terminal-emulator")
			if err != nil || defaultEmulatorName == "" {
				// No emulator config found - not an error, just skip
				slog.Debug("no terminal emulator configuration found",
					"workspaceEmulator", workspaceEmulatorName,
					"workspace", workspaceName,
					"default", "not set")
				return nil
			}
			emulatorDB, err = ds.GetTerminalEmulator(defaultEmulatorName)
			if err != nil {
				return fmt.Errorf("default terminal emulator '%s' not found: %w", defaultEmulatorName, err)
			}
		}
	}

	// 3. Check if it's a wezterm emulator
	if emulatorDB.Type != "wezterm" {
		slog.Debug("terminal emulator is not wezterm type", "name", emulatorDB.Name, "type", emulatorDB.Type)
		return nil
	}

	// 4. Parse the configuration from JSON to WezTerm struct
	config, err := emulatorDB.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to parse emulator config: %w", err)
	}

	// Create WezTerm configuration
	weztermConfig := &wezterm.WezTerm{
		Name:        emulatorDB.Name,
		Description: emulatorDB.Description.String,
		Workspace:   workspaceName,
		Enabled:     emulatorDB.Enabled,
	}

	// Map JSON config to WezTerm struct fields
	if err := mapConfigToWezTerm(config, weztermConfig); err != nil {
		return fmt.Errorf("failed to map config to WezTerm struct: %w", err)
	}

	// 5. Use wezterm.LuaGenerator to generate config
	generator := wezterm.NewLuaGenerator()
	luaConfig, err := generator.GenerateFromConfig(weztermConfig)
	if err != nil {
		return fmt.Errorf("failed to generate wezterm lua config: %w", err)
	}

	// 6. Write to stagingDir/.wezterm.lua
	weztermPath := filepath.Join(stagingDir, ".wezterm.lua")
	if err := os.WriteFile(weztermPath, []byte(luaConfig), 0644); err != nil {
		return fmt.Errorf("failed to write wezterm config: %w", err)
	}

	slog.Debug("generated wezterm config", "name", emulatorDB.Name, "path", weztermPath)
	return nil
}

// mapConfigToWezTerm maps a generic config map to WezTerm struct fields
func mapConfigToWezTerm(config map[string]any, wt *wezterm.WezTerm) error {
	// Set defaults
	wt.Font = wezterm.FontConfig{
		Family: "MesloLGS Nerd Font Mono",
		Size:   14,
	}
	wt.Window = wezterm.WindowConfig{
		Opacity: 1.0,
	}

	// Map font configuration
	if fontConfig, ok := config["font"].(map[string]any); ok {
		if family, ok := fontConfig["family"].(string); ok {
			wt.Font.Family = family
		}
		if size, ok := fontConfig["size"].(float64); ok {
			wt.Font.Size = size
		} else if sizeInt, ok := fontConfig["size"].(int); ok {
			wt.Font.Size = float64(sizeInt)
		}
	}

	// Map window configuration
	if windowConfig, ok := config["window"].(map[string]any); ok {
		if opacity, ok := windowConfig["opacity"].(float64); ok {
			wt.Window.Opacity = opacity
		}
		if blur, ok := windowConfig["blur"].(int); ok {
			wt.Window.Blur = blur
		} else if blurFloat, ok := windowConfig["blur"].(float64); ok {
			wt.Window.Blur = int(blurFloat)
		}
		if decorations, ok := windowConfig["decorations"].(string); ok {
			wt.Window.Decorations = decorations
		}
		if initialRows, ok := windowConfig["initialRows"].(int); ok {
			wt.Window.InitialRows = initialRows
		} else if initialRowsFloat, ok := windowConfig["initialRows"].(float64); ok {
			wt.Window.InitialRows = int(initialRowsFloat)
		}
		if initialCols, ok := windowConfig["initialCols"].(int); ok {
			wt.Window.InitialCols = initialCols
		} else if initialColsFloat, ok := windowConfig["initialCols"].(float64); ok {
			wt.Window.InitialCols = int(initialColsFloat)
		}
		if closeOnExit, ok := windowConfig["closeOnExit"].(string); ok {
			wt.Window.CloseOnExit = closeOnExit
		}
		// Padding
		if paddingLeft, ok := windowConfig["paddingLeft"].(int); ok {
			wt.Window.PaddingLeft = paddingLeft
		} else if paddingLeftFloat, ok := windowConfig["paddingLeft"].(float64); ok {
			wt.Window.PaddingLeft = int(paddingLeftFloat)
		}
		if paddingRight, ok := windowConfig["paddingRight"].(int); ok {
			wt.Window.PaddingRight = paddingRight
		} else if paddingRightFloat, ok := windowConfig["paddingRight"].(float64); ok {
			wt.Window.PaddingRight = int(paddingRightFloat)
		}
		if paddingTop, ok := windowConfig["paddingTop"].(int); ok {
			wt.Window.PaddingTop = paddingTop
		} else if paddingTopFloat, ok := windowConfig["paddingTop"].(float64); ok {
			wt.Window.PaddingTop = int(paddingTopFloat)
		}
		if paddingBottom, ok := windowConfig["paddingBottom"].(int); ok {
			wt.Window.PaddingBottom = paddingBottom
		} else if paddingBottomFloat, ok := windowConfig["paddingBottom"].(float64); ok {
			wt.Window.PaddingBottom = int(paddingBottomFloat)
		}
	}

	// Map color configuration
	if colors, ok := config["colors"].(map[string]any); ok {
		colorConfig := &wezterm.ColorConfig{}

		if fg, ok := colors["foreground"].(string); ok {
			colorConfig.Foreground = fg
		}
		if bg, ok := colors["background"].(string); ok {
			colorConfig.Background = bg
		}
		if cursorBg, ok := colors["cursor_bg"].(string); ok {
			colorConfig.CursorBg = cursorBg
		}
		if cursorFg, ok := colors["cursor_fg"].(string); ok {
			colorConfig.CursorFg = cursorFg
		}
		if cursorBorder, ok := colors["cursor_border"].(string); ok {
			colorConfig.CursorBorder = cursorBorder
		}
		if selBg, ok := colors["selection_bg"].(string); ok {
			colorConfig.SelectionBg = selBg
		}
		if selFg, ok := colors["selection_fg"].(string); ok {
			colorConfig.SelectionFg = selFg
		}

		// ANSI colors (8 colors)
		if ansi, ok := colors["ansi"].([]any); ok {
			ansiColors := make([]string, 0, 8)
			for _, c := range ansi {
				if colorStr, ok := c.(string); ok {
					ansiColors = append(ansiColors, colorStr)
				}
			}
			colorConfig.ANSI = ansiColors
		}

		// Bright colors (8 colors)
		if brights, ok := colors["brights"].([]any); ok {
			brightColors := make([]string, 0, 8)
			for _, c := range brights {
				if colorStr, ok := c.(string); ok {
					brightColors = append(brightColors, colorStr)
				}
			}
			colorConfig.Brights = brightColors
		}

		wt.Colors = colorConfig
	}

	// Map theme reference
	if themeRef, ok := config["themeRef"].(string); ok {
		wt.ThemeRef = themeRef
	}

	// Map scrollback
	if scrollback, ok := config["scrollback"].(int); ok {
		wt.Scrollback = scrollback
	} else if scrollbackFloat, ok := config["scrollback"].(float64); ok {
		wt.Scrollback = int(scrollbackFloat)
	}

	// Map leader key
	if leader, ok := config["leader"].(map[string]any); ok {
		leaderKey := &wezterm.LeaderKey{}
		if key, ok := leader["key"].(string); ok {
			leaderKey.Key = key
		}
		if mods, ok := leader["mods"].(string); ok {
			leaderKey.Mods = mods
		}
		if timeout, ok := leader["timeout"].(int); ok {
			leaderKey.Timeout = timeout
		} else if timeoutFloat, ok := leader["timeout"].(float64); ok {
			leaderKey.Timeout = int(timeoutFloat)
		}
		wt.Leader = leaderKey
	}

	// Map key bindings
	if keys, ok := config["keys"].([]any); ok {
		keybindings := make([]wezterm.Keybinding, 0, len(keys))
		for _, k := range keys {
			if keyMap, ok := k.(map[string]any); ok {
				keybinding := wezterm.Keybinding{}
				if key, ok := keyMap["key"].(string); ok {
					keybinding.Key = key
				}
				if mods, ok := keyMap["mods"].(string); ok {
					keybinding.Mods = mods
				}
				if action, ok := keyMap["action"].(string); ok {
					keybinding.Action = action
				}
				if args, ok := keyMap["args"]; ok {
					keybinding.Args = args
				}
				keybindings = append(keybindings, keybinding)
			}
		}
		wt.Keys = keybindings
	}

	// Map tab bar configuration
	if tabBar, ok := config["tabBar"].(map[string]any); ok {
		tabBarConfig := &wezterm.TabBarConfig{}
		if enabled, ok := tabBar["enabled"].(bool); ok {
			tabBarConfig.Enabled = enabled
		}
		if position, ok := tabBar["position"].(string); ok {
			tabBarConfig.Position = position
		}
		if maxWidth, ok := tabBar["maxWidth"].(int); ok {
			tabBarConfig.MaxWidth = maxWidth
		} else if maxWidthFloat, ok := tabBar["maxWidth"].(float64); ok {
			tabBarConfig.MaxWidth = int(maxWidthFloat)
		}
		if showNewTab, ok := tabBar["showNewTab"].(bool); ok {
			tabBarConfig.ShowNewTab = showNewTab
		}
		if fancyTabBar, ok := tabBar["fancyTabBar"].(bool); ok {
			tabBarConfig.FancyTabBar = fancyTabBar
		}
		if hideTabBarIfOnly, ok := tabBar["hideTabBarIfOnly"].(bool); ok {
			tabBarConfig.HideTabBarIfOnly = hideTabBarIfOnly
		}
		wt.TabBar = tabBarConfig
	}

	// Map pane configuration
	if pane, ok := config["pane"].(map[string]any); ok {
		paneConfig := &wezterm.PaneConfig{}
		if inactiveSat, ok := pane["inactiveSaturation"].(float64); ok {
			paneConfig.InactiveSaturation = inactiveSat
		}
		if inactiveBright, ok := pane["inactiveBrightness"].(float64); ok {
			paneConfig.InactiveBrightness = inactiveBright
		}
		wt.Pane = paneConfig
	}

	return nil
}
