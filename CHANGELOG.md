# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- More comprehensive example projects
- Performance benchmark suite
- CONTRIBUTING.md guide
- Code linting configuration (.golangci.yml)
- Coverage reporting in CI

### Future Considerations
- Interactive scaffold mode with wizards
- Code generation subcommands (generate api/model/crud)
- Microservice architecture templates
- Monitoring integration (Prometheus, Grafana)
- Internationalization support
- Additional middleware (rate limiting, caching, etc.)

## [v0.1.3] - 2026-11-04

### Added

- **Unit tests** for JWT, config loading/validation, errno, CORS, request ID,
  Authn, response, server application lifecycle, MySQL DSN, DB close and utilities.
- **Scaffold end-to-end verification**: `scripts/verify-scaffold.sh` generates a
  project, runs `go mod tidy`/`build`/`vet`, starts it and checks `/health` and `/healthz`.
- **GitHub Actions CI** (`.github/workflows/ci.yml`): build, vet, gofmt and tests.
- **Release notes** for v0.1.2 and v0.1.3 (`RELEASE_v0.1.2.md`, `RELEASE_v0.1.3.md`).
- Server config keys: `shutdown-timeout`, `read-header-timeout`, `read-timeout`,
  `write-timeout`, `idle-timeout`, `max-header-bytes` (defaults 10s/30s/60s/120s/1MB).
- Redis config keys: `pool-size`, `dial-timeout`, `read-timeout`, `write-timeout`.
- JWT `jti` claim on every token; user-level token revocation (`FailureAllClient`).
- `jwt.blacklist-fail-closed` option (default `false`, fail-open with startup warning
  when Redis is not configured).
- Built-in `GET /healthz` liveness endpoint.
- `App.GetEngine()` accessor for custom middleware registration.

### Changed

- `token.Init(opts *Options) error` now validates a Secret of at least 32 bytes,
  stores config atomically, and is called automatically by `server.NewApplication`.
- JWT package name is `token` (import path `.../core/auth/jwt`); `init`/`Sign`/
  `ParseRequest` now return distinguishable errors; `ExpiresIn` is remaining seconds.
- HTTP 状态码恢复语义：401（未认证/Token 错误）、404、429、500；业务码与 HTTP 码分离。
- `errno.Decode` 对未知错误返回统一 500 文案，不再回显 `err.Error()`。
- `middleware.Authn()` must be mounted explicitly on protected routes.
- Request ID is now also stored in `r.Context()` (`log.C(ctx)` works) and validated.
- CORS adds `PATCH` and `X-Requested-With`, echoes whitelist origin with credentials.
- MySQL DSN uses `loc=UTC` and explicit `timeout/readTimeout/writeTimeout`;
  pool defaults and `idle <= open` clamping are applied.
- All built-in config keys support `FRAMEWORK_*` env overrides; config file is optional
  when `Load("")` is used with env variables/defaults.
- `log.Init()` returns an error, rotation failures fail startup, `log.Close()` is
  reference-counted and `log.C(ctx)` attaches request_id.
- Scaffold generator: `go 1.25.1`, requires framework `v0.1.3`, generates runnable
  `config.yaml` + `config.example.yaml`, `/health` route, `scaffold version` command;
  removed the unused viper dependency and template files.
- Makefile: `install` respects `GOBIN`, `test-scaffold` calls the verification script,
  added `release-tag` guidance target; ldflags unchanged.

### Fixed

- `log.Warning` no longer uses the wrong zap method.
- `util.ParamToInt64` reads the URL path parameter instead of the query string.
- `util.GenerateRandomString` uses `crypto/rand` and handles non-positive lengths.
- Duplicate/scattered validation removed: `mysql` is only required when `host` is set.
- `Application.Close()` closes DB before log and is idempotent (`sync.Once`).
- Unified 404 and panic responses (`response.Error` + no internal detail leak).
- Scaffold no longer emits code that references non-existent APIs.

### Security

- JWT: HS256-only verification, required `exp`, ≥32-byte secret, atomic config.
- Request-ID header injection limited to `[A-Za-z0-9._-]{1,128}`.
- SQL filter field names validated (fail-closed `1 = 0`).

### Breaking Changes

- JWT secret must be ≥32 bytes; short secrets now fail `config.Validate()`/`token.Init()`.
- Existing tokens without `jti` are rejected — clients must re-login after upgrade.
- errno responses no longer always use HTTP 200 (clients relying on that must adapt).
- `log.Init` returns an error; `log.Close` refcounts shared writers.
- Generated projects require the `v0.1.3` tag and use `RouteRegisterFunc func(*gin.Engine, *gorm.DB)`.
- `validator` package renamed from `tools`; update imports if you used the old path.

## [v0.1.2] - 2026-08-11

### Added

- **Log rotation**: `log.RotateConfig` supports `daily`, `hourly` and `auto` interval,
  optional stdout mirroring, and files named `${BasePath}.YYYYMMDD.log` (or `YYYYMMDDHH.log`).
- Release notes for v0.1.2 and upgrade guide.

### Changed

- `log.Init` accepts `Rotation *RotateConfig`; `Dir` shortcuts are replaced by
  `OutputPaths`/`Rotation` (see `examples/config.example.yaml`).

### Fixed

- Rotation writer is flushed/closed on `log.Close()`, avoiding truncated logs during shutdown.

## [v0.1.1] - 2026-04-16

### Added

#### Documentation
- **Configuration Usage Guide** - Complete guide for loading and using configurations (CONFIG_USAGE.md)
- **Custom Config Reading** - Comprehensive guide for reading custom config options (CUSTOM_CONFIG.md)
- **Quick Reference Cards** - Quick reference for configuration and custom config (CONFIG_QUICK_REFERENCE.md, CUSTOM_CONFIG_QUICK_REF.md)
- **README Updates** - Added configuration section with examples and documentation links

#### Features
- **Viper Integration** - Scaffold now includes viper for reading custom configurations
- **Custom Config Support** - Users can now read any custom config from config.yaml using viper

### Changed

- Updated scaffold to import `github.com/spf13/viper` in generated projects
- Updated scaffold go.mod to include viper dependency
- Improved main.go template with custom config reading examples
- Enhanced README.md with comprehensive configuration documentation

### Fixed

- Fixed module path issues in all source files
- Fixed scaffold generator to use correct GitHub module path

## [v0.1.0-alpha] - 2026-04-16

### Added

#### Core Features
- **Configuration Management** - Viper-based config loader with YAML support and environment variables
- **Application Server** - HTTP server with graceful shutdown and signal handling
- **Database Layer** - MySQL and Redis connections with connection pooling
- **Authentication** - JWT token management with signing, validation, and blacklist
- **Password Security** - bcrypt-based password encryption
- **Logging System** - Structured logging with Zap, file rotation, and request ID tracking
- **Middleware Suite** - Auth, CORS, RequestID, Logger, and Recovery middleware
- **Response Handling** - Unified JSON response format with success/error helpers
- **Error Management** - Comprehensive error code system
- **Validation** - Data validation utilities with custom regex support
- **Utility Functions** - Time, file, IP, cache, tree, and conversion utilities
- **Version Management** - Build-time version injection and display
- **Service Interfaces** - Abstract interfaces for SMS, Storage, and LLM services

#### Developer Tools
- **Scaffold Tool** - CLI tool for quick project initialization (`scaffold new <project-name>`)
- **Makefile** - Simplified build, install, test, and clean commands
- **Verification Script** - Automated testing script for scaffold tool

#### Documentation
- **README.md** - Comprehensive project overview and quick start guide
- **Quick Start** - Step-by-step tutorial for getting started in 5 minutes
- **Architecture** - Detailed architecture design and best practices
- **Scaffold Guide** - Complete usage documentation for the scaffold tool
- **Development Docs** - Technical documentation for contributors
- **Release Checklist** - Pre-release verification checklist

#### Examples
- **Simple API Example** - Basic example demonstrating framework usage
- **Configuration Example** - Complete config.yaml template

### Changed
- N/A (initial release)

### Fixed
- Import path conflicts between standard library and custom packages
- Package naming inconsistencies in JWT module
- Configuration structure definition issues
- Middleware reference errors

### Security
- JWT secret validation in configuration
- SQL injection prevention through GORM parameterized queries
- Password hashing with bcrypt (automatic salting)
- Token blacklist mechanism for logout functionality

---

**Note**: This project follows Semantic Versioning; check [RELEASE_v0.1.3.md](RELEASE_v0.1.3.md)
for upgrade notes. Please report any issues or suggestions via GitHub Issues!
