# CHANGELOG

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Traffic Monitoring Service**
  - Real-time traffic monitoring (upload/download speeds per interface)
  - Historical traffic data query (up to 7 days)
  - TrafficService interface and implementation
  - TrafficRealtimeItem and TrafficHistoryItem type definitions
  - Input validation for query parameters
- **CRUD Operations for Firewall Service**
  - Add/Edit/Delete ACL rules
  - Add/Edit/Delete DNAT (port forwarding) rules
  - Add/Edit/Delete connection limit rules

- **CRUD Operations for VPN Service**
  - Add/Edit/Delete PPTP clients
  - Add/Edit/Delete L2TP clients

- **CRUD Operations for Network Service**
  - Add/Edit/Delete DNS static entries
  - Add/Edit/Delete static routes
  - Add/Edit/Delete DHCP static bindings

- **Utility Functions** (`utils` package)
  - IP/MAC/Port validation functions
  - `IsValidIPv4()`, `IsValidMAC()`, `IsValidPort()`, `IsValidCIDR()`
  - Retry mechanism with fixed delay: `Retry()`
  - Retry with exponential backoff: `RetryWithBackoff()`
  - Rate limiter: `RateLimiter`
  - Validation helpers: `ValidateIPRange()`, `ValidateMACAddress()`, `ValidatePort()`

- **Comprehensive Test Coverage**
  - Utils package: 100% coverage with 15+ test cases
  - Service layer tests for Monitor, System, Network, Firewall
  - Integration test framework for real router testing

- **Documentation**
  - Advanced example demonstrating all major features
  - Comprehensive utility functions documentation
  - Error handling examples

### Changed
- Enhanced error handling with detailed error messages
- Improved validation for VM service inputs
- Updated all service interfaces to include CRUD methods
- Added godoc comments for better documentation

### Fixed
- Security: Removed all hardcoded credentials from codebase
- Security: Rewrote Git history to remove sensitive information
- Fixed compilation errors in service implementations
- Fixed field references in firewall tests

## [1.0.0] - 2026-03-06

### Added
- Initial SDK implementation with reqv3 HTTP client
- Core client with authentication and session management
- Service layer architecture with interface abstraction
- Monitor service (LAN devices, interfaces, system stats, ARP)
- System service (homepage, upgrade info, backups, web users)
- Network service (WAN, LAN, VLAN, IPv6, IPTV, DDNS, DHCP)
- Firewall service (ACL, DNAT, connection limits)
- VPN service (PPTP and L2TP clients)
- Docker service (images, containers, networks, compose)
- VM service (QEMU virtual machines)
- UPnP service
- Log service
- Automatic version detection (v3/v4)
- Context support with timeout and cancellation
- Comprehensive error handling with error codes
- Cookie-based session management

### Documentation
- Complete README with architecture overview
- Installation and quick start guide
- Service layer documentation
- Error handling guide
- Testing instructions

---

## Version History

- **v1.0.0** (2026-03-06): Initial release with core functionality
- **Unreleased**: CRUD operations, utility functions, enhanced testing
