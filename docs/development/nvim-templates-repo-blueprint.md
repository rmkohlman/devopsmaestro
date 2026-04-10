# Neovim Templates Repository — Planning Document

> **Internal planning document.** This file describes a proposed `devopsmaestro-nvim-templates` repository that has not yet been created. It is kept here for historical reference.

## Proposal Summary

A dedicated repository of production-ready Neovim configurations, curated for developers and DevOps engineers. Templates would be installable directly via `nvp` or by manual clone.

## Planned Template Categories

| Template | Target User |
|----------|-------------|
| `starter` | Most developers — LSP, Telescope, Treesitter, Git |
| `devops` | DevOps/SRE — Terraform, Kubernetes, Docker, Ansible |
| `ide-complete` | Full IDE replacement — DAP debugging, testing, advanced LSP |
| `workspace-optimized` | Container-based dev — lightweight, fast startup |
| `languages/golang` | Go development |
| `languages/python` | Python development |
| `languages/javascript` | JavaScript/TypeScript development |

## Planned CLI Integration

```bash
# Browse available templates
dvm nvim init custom --git-url https://github.com/rmkohlman/devopsmaestro-nvim-templates.git

# Future shorthand (not yet implemented)
dvm nvim templates browse
```

## Status

This repository has not been created. Track progress via the [GitHub Issue Tracker](https://github.com/rmkohlman/devopsmaestro/issues).
