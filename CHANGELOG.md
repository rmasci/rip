# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.1.0] - 2026-01-18

### Added
- Business Source License (BSL 1.1) for commercial use protection while allowing free personal use
- Enhanced README documentation with detailed Plex/Jellyfin library setup instructions
- New "Organizing Categories in Plex/Jellyfin" section explaining how the `-c` category flag maps to directories
- Step-by-step guides for setting up separate media libraries in both Plex and Jellyfin
- Visual directory structure example showing how categories organize files
- `release` target in Makefile to build all platforms and create release binaries
- FileBot pricing information in dependencies section ($6/year or $50 lifetime)
- `release/` directory to .gitignore

### Changed
- Reorganized README structure: moved "Setting Up MergerFS" section after Install section
- Enhanced `-c, --category` parameter documentation to clarify directory mapping
- Updated MergerFS section with cross-references to detailed setup instructions
- Updated repository references from `dvdrip` to `rip` in clone commands

### Technical
- Added RELEASE_DIR variable to Makefile for release builds
- Release target copies binaries from `binaries/<os>/rip` to `release/rip-<os>` with appropriate naming

---

## Format Notes

This changelog follows the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format.

Sections used:
- **Added** for new features
- **Changed** for changes in existing functionality
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** for security-related fixes
