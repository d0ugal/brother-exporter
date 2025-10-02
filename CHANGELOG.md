# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [1.1.5](https://github.com/d0ugal/brother-exporter/compare/v1.1.4...v1.1.5) (2025-10-02)


### Bug Fixes

* enable indirect dependency updates in renovate config ([a9293d1](https://github.com/d0ugal/brother-exporter/commit/a9293d1f2cd731890cc2cd78b85f83a34cbea0ce))

## [1.1.4](https://github.com/d0ugal/brother-exporter/compare/v1.1.3...v1.1.4) (2025-09-22)


### Bug Fixes

* **lint:** whitespace ([184754a](https://github.com/d0ugal/brother-exporter/commit/184754a1896d4cb80ed7854e2a80830d0e00794f))

## [1.1.3](https://github.com/d0ugal/brother-exporter/compare/v1.1.2...v1.1.3) (2025-09-22)


### Bug Fixes

* **lint:** resolve godoclint and gosec issues ([d80a317](https://github.com/d0ugal/brother-exporter/commit/d80a317e14f6dd415f247ad549a9616c68145244))
* **lint:** resolve godoclint and gosec issues ([c3c0677](https://github.com/d0ugal/brother-exporter/commit/c3c0677794e78646ddea55ac0e22284a9f755f5d))

## [1.1.2](https://github.com/d0ugal/brother-exporter/compare/v1.1.1...v1.1.2) (2025-09-20)


### Bug Fixes

* **lint:** resolve gosec configuration contradiction ([b88cf5d](https://github.com/d0ugal/brother-exporter/commit/b88cf5d816f0db32db2b196b7f9fe17b08cd2aab))

## [1.1.1](https://github.com/d0ugal/brother-exporter/compare/v1.1.0...v1.1.1) (2025-09-20)


### Bug Fixes

* **deps:** update module github.com/gin-gonic/gin to v1.11.0 ([5ed3bce](https://github.com/d0ugal/brother-exporter/commit/5ed3bce17522233b67259f1662d820882106028b))
* **deps:** update module github.com/gin-gonic/gin to v1.11.0 ([c8def93](https://github.com/d0ugal/brother-exporter/commit/c8def930ca92cf0a72a9c70af3ca19f955c82e62))

## [1.1.0](https://github.com/d0ugal/brother-exporter/compare/v1.0.4...v1.1.0) (2025-09-12)


### Features

* add printer uptime metric ([fbab094](https://github.com/d0ugal/brother-exporter/commit/fbab09442fdfd692d4340d0d4fd9a8b782df5f62))
* change uptime metric to Unix timestamp for better PromQL support ([fa7a3e7](https://github.com/d0ugal/brother-exporter/commit/fa7a3e7e3271154cbb7cb6bdd8e6dd800ca508af))
* implement proper page counter parsing from Brother counters data ([903767f](https://github.com/d0ugal/brother-exporter/commit/903767fe40310b229e1c33b8c0ccb4fd52d0def3))
* populate brother_printer_info metric with Brother-specific OIDs ([85cf77b](https://github.com/d0ugal/brother-exporter/commit/85cf77b3ea4028c2edc3203360c81b4d357ede1a))


### Bug Fixes

* remove _count suffix from gauge metrics to follow Prometheus best practices ([f3b2a2c](https://github.com/d0ugal/brother-exporter/commit/f3b2a2c0a2b1464bbee77b585c223ee66375c86c))
* remove binary ([ae98e60](https://github.com/d0ugal/brother-exporter/commit/ae98e60043d6571c431e0b6bafa44fccdb04bc2d))

## [1.0.4](https://github.com/d0ugal/brother-exporter/compare/v1.0.3...v1.0.4) (2025-09-12)


### Bug Fixes

* **deps:** update module github.com/gosnmp/gosnmp to v1.42.1 ([41b924a](https://github.com/d0ugal/brother-exporter/commit/41b924a98770d5a568cc8f8d898e2fc1a5eaadbd))
* **deps:** update module github.com/gosnmp/gosnmp to v1.42.1 ([14b97c7](https://github.com/d0ugal/brother-exporter/commit/14b97c77db64ff4f80492d95d846500fb1b446f0))
* **deps:** update module github.com/prometheus/client_golang to v1.23.2 ([333cae2](https://github.com/d0ugal/brother-exporter/commit/333cae2c10846652cd94aee256a3f0f8033f84f9))
* **deps:** update module github.com/prometheus/client_golang to v1.23.2 ([ff4a5a0](https://github.com/d0ugal/brother-exporter/commit/ff4a5a0352a048439aac91a81ec6dba9a5dd6146))
* **deps:** update module github.com/stretchr/testify to v1.11.1 ([ab18d5b](https://github.com/d0ugal/brother-exporter/commit/ab18d5b9715b187f398b5617b835b504c0f69a67))
* **deps:** update module github.com/stretchr/testify to v1.11.1 ([88066ce](https://github.com/d0ugal/brother-exporter/commit/88066cead413e82ab3f1fbea4a3faaa21b2a426f))

## [1.0.3](https://github.com/d0ugal/brother-exporter/compare/v1.0.2...v1.0.3) (2025-09-11)


### Bug Fixes

* correct page count collection to use standard MIB OIDs ([1ed8232](https://github.com/d0ugal/brother-exporter/commit/1ed8232f63e51954a514734c834a9b2aa38b949f))
* ignore the binary ([c099634](https://github.com/d0ugal/brother-exporter/commit/c099634fbc76a25a275818ed4da79ee978ca30f6))

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
