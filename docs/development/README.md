# Development Documentation

This folder contains documentation about DevOpsMaestro's development process, architecture decisions, and release workflows.

---

## 📚 Contents

### [decisions.md](./decisions.md)
**Architecture Decision Records (ADRs)** - Major technical decisions and their rationale.

Documents why we chose specific approaches:
- Theme system design
- Release notes format
- Licensing model
- Session state documentation
- YAML syntax highlighting
- Binary distribution

### [release-process.md](./release-process.md)
**Release Workflow** - Step-by-step guide for creating releases.

Covers:
- Version numbering (semantic versioning)
- Building cross-platform binaries
- Creating GitHub releases
- Updating documentation
- Verification steps

### [v0.2.0-theme-system.md](./v0.2.0-theme-system.md)
**Theme System Development Notes** - Detailed implementation notes for v0.2.0.

Includes:
- Implementation details
- Design decisions
- Testing strategy
- Performance considerations
- Future enhancements

---

## 🎯 Purpose

This documentation serves multiple purposes:

1. **Transparency** - Shows our development process publicly
2. **Education** - Helps contributors understand decisions
3. **Continuity** - Maintains project knowledge over time
4. **Quality** - Encourages thoughtful decision-making
5. **Collaboration** - Makes it easier for others to contribute

---

## 🤝 Contributing to Development Docs

When making significant technical decisions:

1. Document the decision in `decisions.md` as an ADR
2. Include context, alternatives considered, and rationale
3. Note consequences (pros and cons)
4. Reference related issues or pull requests

When implementing major features:

1. Create a feature-specific markdown file (e.g., `v0.X.0-feature-name.md`)
2. Document implementation details
3. Include testing notes
4. Note future enhancement ideas

---

## 📖 Related Documentation

- **[README.md](../../README.md)** - Project overview and quick start
- **[CHANGELOG.md](../../CHANGELOG.md)** - Version history
- **[LICENSING.md](../../LICENSING.md)** - License information
- **[.ai-session/](../../.ai-session/)** - AI session context (local only, not in repo)

---

## 🔍 For More Information

- **GitHub Repository:** https://github.com/rmkohlman/devopsmaestro
- **Issue Tracker:** https://github.com/rmkohlman/devopsmaestro/issues
- **Discussions:** https://github.com/rmkohlman/devopsmaestro/discussions

---

**Last Updated:** 2026-01-24 (v0.2.0)
