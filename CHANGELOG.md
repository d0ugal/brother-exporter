# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [1.10.1](https://github.com/d0ugal/brother-exporter/compare/v1.10.0...v1.10.1) (2025-10-25)


### Bug Fixes

* Update module github.com/d0ugal/promexporter to v1.4.1 ([f69dd9e](https://github.com/d0ugal/brother-exporter/commit/f69dd9e114e230d6986c95572e648189df480941))
* Update module github.com/prometheus/procfs to v0.19.0 ([9645973](https://github.com/d0ugal/brother-exporter/commit/96459730138ba5c1509e5ae644de56b94ba4c353))

## [1.10.0](https://github.com/d0ugal/brother-exporter/compare/v1.9.0...v1.10.0) (2025-10-25)


### Features

* update promexporter to v1.4.0 ([b23cb04](https://github.com/d0ugal/brother-exporter/commit/b23cb048aca8e69ef054dfbe7ed79db16d41ef8e))


### Bug Fixes

* Update module github.com/d0ugal/promexporter to v1.1.0 ([93e026c](https://github.com/d0ugal/brother-exporter/commit/93e026c2adc4feb3ddd2ef4325b9c6f3c76317d6))
* Update module github.com/d0ugal/promexporter to v1.2.0 ([0936bcc](https://github.com/d0ugal/brother-exporter/commit/0936bcc11f7bc39857191a7b7dc7041ac6881ac2))
* Update module github.com/d0ugal/promexporter to v1.3.1 ([bb627a2](https://github.com/d0ugal/brother-exporter/commit/bb627a2d841b0e11394a359c6bbaf23dcf434268))

## [1.9.0](https://github.com/d0ugal/brother-exporter/compare/v1.8.0...v1.9.0) (2025-10-22)


### Features

* migrate brother-exporter to promexporter library ([ee24e4f](https://github.com/d0ugal/brother-exporter/commit/ee24e4f66ab0609dd4b78fdd717f25979bd4a5d6))
* update to promexporter v1.0.0 ([b1db6b7](https://github.com/d0ugal/brother-exporter/commit/b1db6b76ea92e41d9cb40c179bf8197d21bc2ca1))


### Bug Fixes

* remove problematic config tests to unblock CI ([247b01f](https://github.com/d0ugal/brother-exporter/commit/247b01f4fc6cd829f2c041191bc3c34cfdb4ccf5))
* resolve linting issues ([75955bf](https://github.com/d0ugal/brother-exporter/commit/75955bff1d5a5bbb958da23f8d0465e45c3c6b5d))
* restore correct metric names with printer prefix ([74deed5](https://github.com/d0ugal/brother-exporter/commit/74deed54f0a5605b28a5a1d7553d283f805bcabc))
* restore stable version metric info and labels ([4ab232e](https://github.com/d0ugal/brother-exporter/commit/4ab232e9b2196bf8874008f8caf01c08cd7c410b))
* update config tests to use promexporter structure ([11fdda4](https://github.com/d0ugal/brother-exporter/commit/11fdda42021074e46fbf1546f3e6400d4ea7aa8b))
* update go.sum for promexporter v1.0.0 ([1a5598a](https://github.com/d0ugal/brother-exporter/commit/1a5598a8579e941b24b3ec7dc8b81ed026d6bf42))
* Update module github.com/d0ugal/promexporter to v1 ([ecb0128](https://github.com/d0ugal/brother-exporter/commit/ecb0128cd7b5e09a9be57d6da315e40f97c8fd86))
* Update module github.com/d0ugal/promexporter to v1.0.1 ([c174130](https://github.com/d0ugal/brother-exporter/commit/c174130b3900f423d140177dcea1f7418a32b377))
* Update module github.com/prometheus/procfs to v0.18.0 ([5dc763f](https://github.com/d0ugal/brother-exporter/commit/5dc763f26e28ed0decd52cf913b19a0014b3b740))
* update to latest promexporter changes ([012a512](https://github.com/d0ugal/brother-exporter/commit/012a5121839bc5bd3579bbf285da66a91fc0eb10))

## [1.8.0](https://github.com/d0ugal/brother-exporter/compare/v1.7.1...v1.8.0) (2025-10-14)


### Features

* set Gin to release mode unless debug logging is enabled ([b49cc56](https://github.com/d0ugal/brother-exporter/commit/b49cc560cf899d2d6c5f03db47c00016d52e63fd))


### Bug Fixes

* auto-fix import ordering with golangci-lint ([07a4dda](https://github.com/d0ugal/brother-exporter/commit/07a4dda95842284d42ca4004bb43022d48398d99))
* correct import ordering for gci linter ([cd2daee](https://github.com/d0ugal/brother-exporter/commit/cd2daeeeb5fa9b5f99c4098676573c4ddb6be6d7))

## [1.7.1](https://github.com/d0ugal/brother-exporter/compare/v1.7.0...v1.7.1) (2025-10-14)


### Bug Fixes

* Update dependency go to v1.25.3 ([042a340](https://github.com/d0ugal/brother-exporter/commit/042a340971e8d1b68ef4522ca6bd435dc6b09ada))
* Update module golang.org/x/tools to v0.38.0 ([2e503b3](https://github.com/d0ugal/brother-exporter/commit/2e503b3f01a3bfc43df3bd9e8c7c5b566b8dcbfe))

## [1.7.0](https://github.com/d0ugal/brother-exporter/compare/v1.6.0...v1.7.0) (2025-10-08)


### Features

* update dependencies to v0.22.0 ([41113c6](https://github.com/d0ugal/brother-exporter/commit/41113c6de07134fc507e8b607a0d00fc3a586f10))
* update dependencies to v0.29.0 ([6621411](https://github.com/d0ugal/brother-exporter/commit/6621411682f5f4f61bc4da563177e1c0a4781fb7))
* Update module golang.org/x/crypto to v0.43.0 ([9fe49e5](https://github.com/d0ugal/brother-exporter/commit/9fe49e5167322b81c204fc7d3392343cdbc6180a))
* Update module golang.org/x/sys to v0.37.0 ([5f4d06a](https://github.com/d0ugal/brother-exporter/commit/5f4d06aa9eaa3e19f79734853336fa72bd9510b2))
* Update module golang.org/x/text to v0.30.0 ([971f952](https://github.com/d0ugal/brother-exporter/commit/971f9526d94cfb128eae506217ea695e800daaca))


### Bug Fixes

* update gomod commitMessagePrefix from feat to fix ([f41d0f8](https://github.com/d0ugal/brother-exporter/commit/f41d0f899252b7ad31330f8949c4e0773bf95c84))

## [1.6.0](https://github.com/d0ugal/brother-exporter/compare/v1.5.0...v1.6.0) (2025-10-08)


### Features

* update dependencies to v0.45.0 ([70e9f0e](https://github.com/d0ugal/brother-exporter/commit/70e9f0e8f5f3613daf6de0cfdb4aa284f385636e))
* update dependencies to v1.25.2 ([059bdcc](https://github.com/d0ugal/brother-exporter/commit/059bdcc2c4347ab44d42f336e72736816ab960b2))

## [1.5.0](https://github.com/d0ugal/brother-exporter/compare/v1.4.0...v1.5.0) (2025-10-07)


### Features

* update dependencies to v0.67.1 ([85b0959](https://github.com/d0ugal/brother-exporter/commit/85b0959f90a20cd79cb4cc7c7e1cb5b9f054a679))

## [1.4.0](https://github.com/d0ugal/brother-exporter/compare/v1.3.0...v1.4.0) (2025-10-06)


### Features

* **renovate:** use feat: commit messages for dependency updates ([7d214aa](https://github.com/d0ugal/brother-exporter/commit/7d214aa291c659a07a45616c1d982134ed82803c))
* update dependencies to v2.4.3 ([dd3a893](https://github.com/d0ugal/brother-exporter/commit/dd3a8932788ac82dab10849e9fb9a83a6a8feaed))

## [1.3.0](https://github.com/d0ugal/brother-exporter/compare/v1.2.0...v1.3.0) (2025-10-03)


### Features

* **renovate:** add docs commit message format for documentation updates ([ced03a7](https://github.com/d0ugal/brother-exporter/commit/ced03a7c81304ccd56890613c4ed69ebd34214ca))

## [1.2.0](https://github.com/d0ugal/brother-exporter/compare/v1.1.6...v1.2.0) (2025-10-02)


### Features

* **deps:** migrate to YAML v3 ([4a7f5ff](https://github.com/d0ugal/brother-exporter/commit/4a7f5ff0668a15426be62c11f98bfb70809b5294))
* **renovate:** add gomodTidy post-update option for Go modules ([497b5d2](https://github.com/d0ugal/brother-exporter/commit/497b5d2cfaceaea3519949384e91e53b6b42b37b))

## [1.1.6](https://github.com/d0ugal/brother-exporter/compare/v1.1.5...v1.1.6) (2025-10-02)


### Reverts

* remove unnecessary renovate config changes ([7d595cf](https://github.com/d0ugal/brother-exporter/commit/7d595cfbebde35415d2fabd70c4305f69910484f))

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
