# gokit

A collection of reusable Go packages for building applications.

## Packages

### tunnel

SSH tunnel management for secure connections through bastion hosts.

**Features:**

- Password and SSH key authentication
- Automatic or fixed local port allocation
- Multiple simultaneous connections support
- Known hosts verification (secure) or insecure mode for development
- Clean resource management

**Installation:**
```bash
go get github.com/paulohsilvestre/gokit/tunnel
```