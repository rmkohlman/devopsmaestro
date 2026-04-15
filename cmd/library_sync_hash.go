package cmd

import (
	"fmt"
	"hash"
	"sort"

	nvimpluginlib "github.com/rmkohlman/MaestroNvim/nvimops/library"
	nvimpkglib "github.com/rmkohlman/MaestroNvim/nvimops/package/library"
	terminalemulatorlib "github.com/rmkohlman/MaestroTerminal/terminalops/emulator/library"
	terminalpkglib "github.com/rmkohlman/MaestroTerminal/terminalops/package/library"
	terminalpluginlib "github.com/rmkohlman/MaestroTerminal/terminalops/plugin/library"
	terminalpromptlib "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/library"
	nvimthemelib "github.com/rmkohlman/MaestroTheme/library"
)

// hashNvimPlugins writes a deterministic representation of all embedded
// nvim plugins into the provided hash writer.
func hashNvimPlugins(h hash.Hash) error {
	lib, err := nvimpluginlib.NewLibrary()
	if err != nil {
		return err
	}
	// lib.List() returns plugins sorted by name (deterministic)
	for _, p := range lib.List() {
		fmt.Fprintf(h, "nvim-plugin:%s:%s:%s:%v:%v",
			p.Name, p.Repo, p.Branch, p.Enabled, p.Lazy)
	}
	return nil
}

// hashNvimThemes writes a deterministic representation of all embedded
// nvim themes into the provided hash writer.
func hashNvimThemes(h hash.Hash) error {
	themeInfos, err := nvimthemelib.List()
	if err != nil {
		return err
	}
	// Sort by name for determinism
	sort.Slice(themeInfos, func(i, j int) bool {
		return themeInfos[i].Name < themeInfos[j].Name
	})
	for _, info := range themeInfos {
		theme, err := nvimthemelib.Get(info.Name)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "nvim-theme:%s:%s:%s:%s",
			theme.Name, theme.Plugin.Repo, theme.Plugin.Branch, theme.Style)
	}
	return nil
}

// hashNvimPackages writes a deterministic representation of all embedded
// nvim packages into the provided hash writer.
func hashNvimPackages(h hash.Hash) error {
	lib, err := nvimpkglib.NewLibrary()
	if err != nil {
		return err
	}
	// lib.List() returns packages sorted by name (deterministic)
	for _, p := range lib.List() {
		fmt.Fprintf(h, "nvim-package:%s:%s:%v",
			p.Name, p.Extends, p.Plugins)
	}
	return nil
}

// hashTerminalPrompts writes a deterministic representation of all embedded
// terminal prompts into the provided hash writer.
func hashTerminalPrompts(h hash.Hash) error {
	lib, err := terminalpromptlib.NewPromptLibrary()
	if err != nil {
		return err
	}
	// lib.List() returns prompts sorted by name
	for _, p := range lib.List() {
		fmt.Fprintf(h, "terminal-prompt:%s:%s:%v",
			p.Name, p.Type, p.Enabled)
	}
	return nil
}

// hashTerminalPlugins writes a deterministic representation of all embedded
// terminal plugins into the provided hash writer.
func hashTerminalPlugins(h hash.Hash) error {
	lib, err := terminalpluginlib.NewPluginLibrary()
	if err != nil {
		return err
	}
	// lib.List() returns plugins sorted by name
	for _, p := range lib.List() {
		fmt.Fprintf(h, "terminal-plugin:%s:%s:%s:%v",
			p.Name, p.Repo, p.Manager, p.Enabled)
	}
	return nil
}

// hashTerminalPackages writes a deterministic representation of all embedded
// terminal packages into the provided hash writer.
func hashTerminalPackages(h hash.Hash) error {
	lib, err := terminalpkglib.NewLibrary()
	if err != nil {
		return err
	}
	// lib.List() returns packages sorted by name
	for _, p := range lib.List() {
		fmt.Fprintf(h, "terminal-package:%s:%s:%v:%v",
			p.Name, p.Extends, p.Plugins, p.Prompts)
	}
	return nil
}

// hashTerminalEmulators writes a deterministic representation of all embedded
// terminal emulators into the provided hash writer.
func hashTerminalEmulators(h hash.Hash) error {
	lib, err := terminalemulatorlib.NewEmulatorLibrary()
	if err != nil {
		return err
	}
	emulators := lib.All()
	// Sort by name for determinism
	sort.Slice(emulators, func(i, j int) bool {
		return emulators[i].Name < emulators[j].Name
	})
	for _, e := range emulators {
		fmt.Fprintf(h, "terminal-emulator:%s:%s:%v",
			e.Name, e.Type, e.Enabled)
	}
	return nil
}
