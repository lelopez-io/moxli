# Changelog

All notable changes to moxli will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-10-03

### Added

**Milestone 1: Data Foundation**
- Bookmark data model with Anybox JSON schema compatibility
- URL normalization system for consistent bookmark matching
- Tag normalization to lower-kebab-case format
- Hierarchical tag support (`[["parent", "child"]]`)

**Milestone 2: I/O Pipeline**
- Anybox JSON importer with full metadata preservation
- Anybox HTML importer for flat exports with TAGS attributes
- Firefox HTML importer (Netscape bookmark format with timestamps)
- Safari HTML importer (Netscape format without timestamps)
- Anybox JSON exporter with round-trip compatibility
- Export validation to ensure schema compliance

**Milestone 3: Merge Engine**
- Base-centric merge strategy (base as source of truth)
- Timestamp reconciliation (prefer oldest non-zero date)
- URL-based bookmark matching with normalized URLs
- Respects intentional deletions (doesn't re-add pruned bookmarks)
- Comprehensive merge test coverage

**Milestone 4: Session and TUI**
- Session management with `~/.moxli/session.yaml` persistence
- Session history tracking (merge records, working directory)
- TUI framework using Bubble Tea
- Welcome screen with continue/new session options
- File discovery system with drag-and-drop path support
- Interactive file selection with format auto-detection
- Base file designation system
- Bookmark browser with lipgloss styling
- Vim-style keyboard navigation (j/k, J/K, h/H, g/G)
- Detail preview overlay (space to toggle)
- Search/filter functionality (/)
- URL opening in default browser (enter from detail)
- Consistent keybinding display across all views

### Technical Details

- **Languages**: Go 1.24+
- **Dependencies**:
  - `charmbracelet/bubbletea` - TUI framework
  - `charmbracelet/lipgloss` - Styling
  - `charmbracelet/bubbles` - UI components
  - `golang.org/x/net/html` - HTML parsing
  - `gopkg.in/yaml.v3` - Session state
- **Architecture**:
  - Internal packages: bookmark, importer, exporter, merge, session, tui
  - Clean separation of concerns
  - Interface-based import/export pipeline
  - Comprehensive test coverage

### Design Decisions

- **Base-centric merge**: Base file represents curated truth, sources enhance only
- **Timestamp-only merge**: MVP focuses on date recovery, not metadata merging
- **Session persistence**: Git-like workflow with working directory state
- **TUI-first**: Pure TUI application (no CLI subcommands for MVP)
- **Format agnostic**: Auto-detection with pluggable importers

### Notes

This is the MVP release of moxli, completing all planned milestones for basic
bookmark management workflow. Post-MVP features (duplicate detection, metadata
merging, conflict resolution) are planned for future releases.

[0.1.0]: https://github.com/lelopez-io/moxli/releases/tag/v0.1.0
