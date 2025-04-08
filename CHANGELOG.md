# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2024-04-08

### Added
- Initial release of PSCloud Exporter
- Support for OAuth 2.0 authentication with PKCE
- Metrics for account balance (prepay, credit, debt)
- Metrics for domains (expiry, status, prices)
- Metrics for nameservers (status, IP count)
- Metrics for invoices (total, status, items)
- Scrape metrics (duration, success, errors)
- Support for both .yml and .yaml configuration files
- Graceful shutdown handling
- Comprehensive error handling and logging
- Docker support
- Makefile for common operations

### Changed
- Updated authentication to use PS.KZ's new OAuth 2.0 system
- Improved error handling with detailed error types
- Enhanced metric collection with better failure handling

### Fixed
- Authentication issues with the new PS.KZ API
- Error handling in metric collection
- Configuration file handling 