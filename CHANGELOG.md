# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- **Project Templates** - Ready-to-use templates for biz, controller, and SQL layers
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

### Deprecated
- N/A

### Removed
- N/A

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

### Known Issues
- No unit tests yet (planned for v0.2.0)
- HTTPS server support not implemented (marked as optional)
- No CI/CD pipeline configured
- Limited example projects
- No performance benchmarks available

## [Unreleased]

### Planned for v0.2.0
- Unit tests for core modules (target: >60% coverage)
- HTTPS server support with TLS
- CI/CD pipeline with GitHub Actions
- More comprehensive example projects
- Performance benchmark suite
- CONTRIBUTING.md guide
- Code linting configuration (.golangci.yml)

### Future Considerations
- Interactive scaffold mode with wizards
- Plugin system for custom templates
- Code generation subcommands (generate api/model/crud)
- Microservice architecture templates
- Monitoring integration (Prometheus, Grafana)
- Internationalization support
- Additional middleware (rate limiting, caching, etc.)

---

**Note**: This is an alpha release. APIs may change in future versions. Please report any issues or suggestions!
