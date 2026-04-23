# Theme Colors

DevOpsMaestro uses your active Neovim theme to color CLI output.

## Setting a Theme

```bash
# Set theme for current workspace
dvm set theme coolnight-ocean

# Set theme at different hierarchy levels
dvm set theme coolnight-synthwave --app myapp
dvm set theme coolnight-forest --domain backend
dvm set theme coolnight-midnight --ecosystem my-platform
```

## Disabling Colors

```bash
# Via flag
dvm status --no-color

# Via environment variable
NO_COLOR=1 dvm get apps

# Via terminal type
TERM=dumb dvm get workspaces
```

## Available Themes

Themes are managed by [MaestroTheme](https://rmkohlman.github.io/MaestroTheme/). See the [Theme Hierarchy](https://rmkohlman.github.io/MaestroTheme/configuration/theme-hierarchy/) documentation for how themes cascade through the object hierarchy.
