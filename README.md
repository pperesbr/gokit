# gokit

A collection of reusable Go packages for building applications.

## Packages

| Package | Description |
|---------|-------------|
| [tunnel](pkg/tunnel/README.md) | SSH tunnel management for secure connections through bastion hosts |
| [dsn](pkg/dsn/README.md) | Database connection string builder for Oracle, PostgreSQL and MySQL |

---

### tunnel

SSH tunnel management for secure connections through bastion hosts.

**Features:**
- Password and SSH key authentication
- Automatic or fixed local port allocation
- Multiple simultaneous connections support
- Known hosts verification (secure) or insecure mode for development
- Connection statistics and lifecycle management

[View documentation](pkg/tunnel/README.md)

---

### dsn

Database connection string builder with factory pattern support.

**Features:**
- Oracle support (Standalone, RAC, DataGuard)
- PostgreSQL support
- MySQL support
- YAML configuration with auto-detect
- Validation and error handling

[View documentation](pkg/dsn/README.md)

---

## Installation

```bash
go get github.com/pperesbr/gokit
```
