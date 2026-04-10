# Contributing to DevOpsMaestro

Thank you for your interest in contributing to DevOpsMaestro!

## Getting Started

### Prerequisites

- **Go 1.25+**
- **Git**
- **Docker or Colima** — for testing container features
- **golangci-lint** (optional) — for linting

### Clone and Build

```bash
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro

# Build dvm
go build -o dvm .

# Build nvp
go build -o nvp ./cmd/nvp/

# Verify
./dvm version
./nvp version
```

### Run Tests

```bash
go test ./...
```

---

## Making Changes

### Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### Commit Style

We use [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git commit -m "feat: add new workspace command"
git commit -m "fix: correct plugin deletion error"
git commit -m "docs: update installation guide"
git commit -m "test: add plugin store tests"
git commit -m "chore: update dependencies"
```

### Submit a Pull Request

```bash
git push origin feature/your-feature-name
```

Then open a Pull Request on GitHub.

---

## Code Standards

Before writing code, read:

- **[STANDARDS.md](https://github.com/rmkohlman/devopsmaestro/blob/main/STANDARDS.md)** — Design patterns and coding standards

Key principles:
- **Decoupling** — Interface → Implementation → Factory pattern
- **Single Responsibility** — Each component has one job
- **Testability** — Every interface has a mock

---

## Documentation

When making changes, update documentation as appropriate:

| Change Type | Update |
|-------------|--------|
| New command | README.md, command docs |
| New feature | Feature docs, examples |
| Bug fix | Add regression test |
| Any change | CHANGELOG.md |

Build and preview docs locally:

```bash
pip install mkdocs-material
mkdocs serve
# Visit http://127.0.0.1:8000
```

---

## Release Process

Releases are handled by maintainers. See [Release Process](./release-process.md) for details.

---

## Getting Help

- **Issues:** [GitHub Issues](https://github.com/rmkohlman/devopsmaestro/issues)
- **Discussions:** [GitHub Discussions](https://github.com/rmkohlman/devopsmaestro/discussions)

---

## License

By contributing, you agree that your contributions will be licensed under GPL-3.0.

---

Thank you for contributing to DevOpsMaestro!
