# Changelog

## [Unreleased]

### Added
- **logx.Sanitizable interface** - Automatic secret sanitization in logging
  - Configs implementing `Sanitizable` are automatically sanitized when logged with `logx.Any()`
  - Prevents accidental exposure of passwords, API keys, tokens, DSNs in logs
  - Defense-in-depth security: developers don't need to remember to sanitize manually
  - Zero performance overhead (simple type assertion + reflection for nil check)
  - See examples in `logx/adapter_sanitize_test.go`

### Security
- **Auto-sanitization in logger** - `logx.Any()` now automatically redacts secrets
  - Applies to any config struct implementing `Sanitize() any` method
  - Handles nil pointers gracefully without panics
  - Opt-out available via direct `zap.Any()` for debugging scenarios

## [0.2.1] - 2025-10-31

### Added
- **configx.NewWithReader()** - In-memory YAML configuration loader for tests
  - Load configuration from `io.Reader` without filesystem I/O
  - Same decode hooks, validation, and env var support as production loader
  - Perfect for unit tests - faster, more reliable, better isolation
  - See `docs/TESTING.md` for usage patterns and best practices

### Changed
- None

### Fixed
- None

### Documentation
- Added `docs/TESTING.md` - Comprehensive testing guide for GoStratum modules
- Updated core README with testing section and NewWithReader examples
- Enhanced testing patterns documentation across framework


## [0.2.0] - 2025-10-29

### Added
- Core module: implemented `New` and `Run` helpers to simplify creating an Fx-based application (New Fx app composition helpers).
- Configuration improvements: added `configx` package and a typed configuration loader with validation and environment variable binding.
- Sanitization: added sanitization utilities and tests for sensitive data handling.
- Version management: release/version management scripts and Makefile targets to streamline releases and bumping versions.

### Changed
- Replace previous logger module with `logx` and refactor logging initialization to use `zap.Sugar` where applicable.
- Configuration loader enhancements: loader options, environment variable binding precedence, and file-based test configuration for clearer tests.
- Migrate several internal providers and Fx initialization to separate concerns (logger, config, lifecycle providers).

### Fixed
- Ensure custom time encoder is only set when necessary in `NewLogger`.
- Update Go toolchain version to 1.25.1 in `go.mod` files where applicable.

### Refactored
- Removed `NewSugared` helper from logger; callers now construct `zap.Sugar` explicitly.
- Extracted and moved `FxEventLogger` out of `logger.go` into `module.go` to improve module boundaries.
- Simplified configuration handling by removing legacy `walkFields` utility and replacing certain viper usages in tests with explicit file-based config for clarity and maintainability.

### Tests
- Added integration tests for configuration binding and alias precedence.
- Added unit tests for `configx` and `logx` modules and improved test coverage for logger functionality.

### Docs
- README and deployment documentation updated to document `ENV_PREFIX` and the new release flow.



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
