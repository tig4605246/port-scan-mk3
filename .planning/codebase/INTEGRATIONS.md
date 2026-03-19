# INTEGRATIONS.md - External Integrations

## Pressure API

- **Type**: HTTP REST API
- **Purpose**: External pressure monitoring for auto-pause
- **Endpoints**: 
  - Simple: Single URL returning `{"pressure": N}`
  - Authenticated: Token endpoint + data endpoint with Bearer token
- **Configuration**:
  - `-pressure-api`: API URL
  - `-pressure-interval`: Polling interval (default 5s)
  - `-disable-api`: Disable API polling

## Network

- **TCP Scanning**: Uses `net.DialTimeout` from Go standard library
- **Protocol**: TCP only
- **Timeout**: Configurable (default 100ms)

## File I/O

- **Input**: CSV files (CIDR, ports)
- **Output**: CSV files (scan results)
- **Resume**: JSON state files

## No Database

This is a stateless CLI tool with no persistent database.

## Docker (E2E Testing Only)

- **Purpose**: Isolated network testing
- **Services**: Mock pressure API
- **Run**: `docker compose` for E2E tests
