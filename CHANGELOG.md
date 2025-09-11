# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [1.0.2](https://github.com/d0ugal/brother-exporter/compare/v1.0.1...v1.0.2) (2025-09-11)


### Bug Fixes

* Add default renovate config ([1a2c9cd](https://github.com/d0ugal/brother-exporter/commit/1a2c9cdc8e56c1e86f9365cad76dd429c305d56f))
* remove stale comment ([88c0fd4](https://github.com/d0ugal/brother-exporter/commit/88c0fd4c2b8f9b25a14553895878eb9119c8d635))

## [1.0.1](https://github.com/d0ugal/brother-exporter/compare/v1.0.0...v1.0.1) (2025-09-11)


### Bug Fixes

* delete binary ([e001786](https://github.com/d0ugal/brother-exporter/commit/e001786cab57ebf39c99026b634365b66a690eb7))
* renovate name ([3eee5a5](https://github.com/d0ugal/brother-exporter/commit/3eee5a5dcd5ecc0abc97c6e4eaa6622e2b4a90ec))

## 1.0.0 (2025-09-10)


### Features

* initial Brother printer exporter implementation ([0a6195c](https://github.com/d0ugal/brother-exporter/commit/0a6195c18c78c196b1220352f1eee3ca17c825b8))

## [Unreleased]

### Features

* Initial implementation of Brother printer exporter
* SNMP-based monitoring for Brother printers
* Comprehensive metrics including toner/drum levels, page counts, and maintenance data
* Support for both percentage and page-based life tracking
* HTML dashboard with metrics overview
* Professional UI with version information and configuration display
* Robust error handling and fallback to standard MIB metrics
* Accurate Brother-specific SNMP data collection using proprietary OIDs
* Support for individual color page counters and duplex printing metrics
