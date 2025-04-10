# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Full support for Kubernetes API (k8saas) for collecting metrics about clusters and projects
- Stub implementations for cases when API services are unavailable

### Changed
- Improved error handling mechanism to increase resilience when API changes
- Added fault-tolerant processing of GraphQL requests for K8S, VPS and other APIs

### Fixed
- Fixed errors in requests to Kubernetes API (k8saas)
- Fixed errors in requests to VPS API related to data structure incompatibility
- Added ability to return empty data instead of errors when API is unavailable

## [0.1.0] - 2025-04-10

### Added
- Full support for Kubernetes API (k8saas) for collecting metrics about clusters and projects
- Stub implementations for cases when API services are unavailable

### Changed
- Improved error handling mechanism to increase resilience when API changes
- Added fault-tolerant processing of GraphQL requests for K8S, VPS and other APIs

### Fixed
- Fixed errors in requests to Kubernetes API (k8saas)
- Fixed errors in requests to VPS API related to data structure incompatibility
- Added ability to return empty data instead of errors when API is unavailable

## [0.0.0] - 2025-04-10

### Added
- Support for custom baseURL in API client for testing purposes
- Improved error handling in GraphQL queries
- Better server metrics (RAM, CPU cores, status, IP count)
- Kubernetes (k8saas) metrics support with cluster and node group statistics
- LBaaS (Load Balancer as a Service) metrics for monitoring load balancers, listeners, pools, and members
- Authentication validation at startup (can be skipped with -skip-auth-check flag)
- Better error handling for GraphQL authentication errors
- Support for the latest PS.KZ API changes
- Improved error handling for unavailable API endpoints
- Graceful degradation when domain API is unavailable
- Stub implementations for unsupported API endpoints to improve resilience
- Error handling in case of API schema changes
- Support for the latest PS.KZ API changes
- Improved error handling for unavailable API endpoints
- Graceful degradation when domain API is unavailable

### Changed
- Updated authentication method to use new PS.KZ API authentication
- Improved error handling with more detailed error types
- Translated all Russian comments to English for better code readability
- Enhanced metric collection with better failure handling
- Optimized cloud server information collection
- Removed unused models directory to simplify code organization
- Removed DBaaS endpoint as it's not ready for GraphQL integration
- Adapted authentication to use token-based access instead of user/password
- Updated GraphQL queries to match the current API schema
- Converted all remaining Russian comments to English
- Improved error handling and reporting
- Adapted authentication to use token-based access instead of user/password
- Updated GraphQL queries to match the current API schema
- Converted all remaining Russian comments to English
- Improved error handling and reporting
- Updated all GraphQL queries to be compatible with the latest API
- Improved error handling strategy to gracefully degrade when certain APIs are unavailable
- Enhanced graceful degradation for LBaaS, K8S, Cloud, VPS, and Project APIs


### Fixed
- Authentication issues with the new PS.KZ API
- Error handling in metric collection
- Configuration file parsing issues
- Several linter warnings and code quality issues
- Authentication issues with the new PS.KZ API
- API compatibility issues with domain listing and counters
- Error messages and documentation consistency
- All scrape error issues with various API endpoints
- Compatibility issues with LBaaS, K8S, Cloud, VPS, and Project APIs
- Various error handling mechanisms for better reliability

## [0.0.0] - 2024-04-08

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