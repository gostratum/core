# Changelog

## [Unreleased]

## [0.1.10] - 2025-10-26

### Added
- Release version 0.1.10

### Changed
- Updated gostratum dependencies to latest versions


# Changelog

All notable changes to the `gostratum/core` module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.9] - 2025-10-26

### Added
- **ENV_PREFIX environment variable** - Global environment variable prefix configuration
  - Allows setting default env prefix without code changes via `ENV_PREFIX` environment variable
  - Precedence: `WithEnvPrefix()` option > `ENV_PREFIX` env var > `STRATUM` default
  - Enables deployment scripts to configure prefix globally: `export ENV_PREFIX=MYAPP`

### Changed
- Updated configx loader to check `ENV_PREFIX` environment variable before applying default
- Enhanced README with comprehensive ENV_PREFIX documentation and examples

### Fixed
- None

## [0.1.8] - 2025-10-XX

### Changed
- Logger improvements to preserve development encoder defaults
- Various documentation updates

## [0.1.7] - 2025-10-XX

### Added
- Health check timeout configuration via `STRATUM_HEALTH_TIMEOUT_MS`

## [0.1.6] - 2025-10-XX

### Added
- Initial configx package with typed configuration loading
- Support for struct tag defaults and validation
- Automatic environment variable binding

## [0.1.5] - 2025-10-XX

### Added
- Health registry for liveness and readiness checks
- Application lifecycle management via Uber FX

## Earlier Versions

See git history for changes prior to 0.1.5.

---

## Version Numbering

This module follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions  
- **PATCH** version for backwards-compatible bug fixes

## Deployment Scripts

This CHANGELOG follows a format compatible with automated deployment tools:

```bash
# Extract latest version
VERSION=$(head -n 20 CHANGELOG.md | grep -oP '^\[\K[0-9]+\.[0-9]+\.[0-9]+' | head -1)

# Extract release notes for latest version
sed -n "/## \[$VERSION\]/,/## \[/p" CHANGELOG.md | sed '$d'

# Tag and release
git tag "v$VERSION"
git push origin "v$VERSION"
```

## Migration Guides

### Migrating to 0.1.9 (ENV_PREFIX support)

**No breaking changes.** This is a backward-compatible addition.

**Before:**
```go
loader := configx.New(
    configx.WithEnvPrefix("MYAPP"),  // Only way to customize prefix
)
```

**After (Option 1 - via environment variable):**
```bash
export ENV_PREFIX=MYAPP  # Set globally
```
```go
loader := configx.New()  // Automatically uses MYAPP prefix
```

**After (Option 2 - still using WithEnvPrefix):**
```go
loader := configx.New(
    configx.WithEnvPrefix("MYAPP"),  // Still works, highest priority
)
```

**Deployment Script Example:**
```bash
#!/bin/bash
# deploy.sh
export ENV_PREFIX=MYAPP
export MYAPP_APP_PORT=8080
export MYAPP_DB_HOST=prod-db.example.com
./myapp
```
