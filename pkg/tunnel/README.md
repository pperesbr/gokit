# Tunnel - SSH Port Forwarding

Package for managing SSH tunnels with support for password and key-based authentication, connection statistics, and lifecycle management.

## Installation

```go
import "github.com/pperesbr/gokit/pkg/tunnel"
```

## Features

- Password and SSH key authentication
- Known hosts file validation (secure mode)
- Insecure mode for development/testing
- Connection statistics (bytes in/out, active connections)
- Tunnel lifecycle management (start, stop, restart)
- Thread-safe operations

## Quick Start

```go
import "github.com/pperesbr/gokit/pkg/tunnel"

// Create SSH config
cfg, err := tunnel.NewSSHConfig(
    "user",           // username
    "password",       // password (or empty if using key)
    "",               // key file path (or empty if using password)
    "bastion.com",    // SSH host
    "",               // known_hosts file (empty for insecure mode)
    22,               // SSH port
)
if err != nil {
    log.Fatal(err)
}

// Create tunnel: forward local:5432 -> remote-db:5432
t := tunnel.NewTunnel(cfg, "remote-db.internal", 5432, 5432)

// Start tunnel
if err := t.Start(); err != nil {
    log.Fatal(err)
}
defer t.Close()

// Use the tunnel
fmt.Printf("Tunnel listening on %s\n", t.LocalAddr())
// Connect to localhost:5432 to reach remote-db.internal:5432
```

## Authentication

### Password Authentication

```go
cfg, err := tunnel.NewSSHConfig(
    "user",
    "password",
    "",              // no key file
    "bastion.com",
    "",
    22,
)
```

Password authentication includes both `password` and `keyboard-interactive` methods for compatibility with different SSH servers.

### SSH Key Authentication

```go
cfg, err := tunnel.NewSSHConfig(
    "user",
    "",                      // no password
    "/home/user/.ssh/id_rsa", // key file path
    "bastion.com",
    "",
    22,
)
```

Supports OpenSSH private key format. If both password and key file are provided, key file takes precedence.

## Host Key Verification

### Secure Mode (Recommended for Production)

```go
cfg, err := tunnel.NewSSHConfig(
    "user",
    "password",
    "",
    "bastion.com",
    "/home/user/.ssh/known_hosts", // known_hosts file
    22,
)

if !cfg.IsInsecure() {
    fmt.Println("Using secure host key verification")
}
```

### Insecure Mode (Development/Testing Only)

```go
cfg, err := tunnel.NewSSHConfig(
    "user",
    "password",
    "",
    "bastion.com",
    "",  // empty = insecure mode
    22,
)

if cfg.IsInsecure() {
    fmt.Println("WARNING: Host key verification disabled")
}
```

## Tunnel Lifecycle

### Start

```go
t := tunnel.NewTunnel(cfg, "remote-host", 5432, 5432)

if err := t.Start(); err != nil {
    log.Fatal(err)
}
```

### Stop

```go
if err := t.Stop(); err != nil {
    log.Printf("Error stopping tunnel: %v", err)
}
```

### Restart

```go
if err := t.Restart(); err != nil {
    log.Fatal(err)
}
```

### Close (alias for Stop)

```go
defer t.Close()
```

## Dynamic Port Allocation

Use port `0` to let the system allocate an available port:

```go
t := tunnel.NewTunnel(cfg, "remote-host", 5432, 0) // local port = 0

if err := t.Start(); err != nil {
    log.Fatal(err)
}

fmt.Printf("Tunnel listening on port %d\n", t.LocalPort())
```

## Tunnel Status

```go
switch t.Status() {
case tunnel.StatusStopped:
    fmt.Println("Tunnel is stopped")
case tunnel.StatusStarting:
    fmt.Println("Tunnel is starting")
case tunnel.StatusRunning:
    fmt.Println("Tunnel is running")
case tunnel.StatusError:
    fmt.Printf("Tunnel error: %v\n", t.LastError())
}
```

## Connection Statistics

```go
stats := t.Stats()

fmt.Printf("Bytes received: %d\n", stats.BytesIn)
fmt.Printf("Bytes sent: %d\n", stats.BytesOut)
fmt.Printf("Total connections: %d\n", stats.Connections)
fmt.Printf("Active connections: %d\n", stats.ActiveConnections)
fmt.Printf("Last activity: %v\n", stats.LastActivity)
fmt.Printf("Started at: %v\n", stats.StartedAt)
```

## Useful Methods

```go
// Get local address (e.g., "127.0.0.1:5432")
localAddr := t.LocalAddr()

// Get remote address (e.g., "remote-host:5432")
remoteAddr := t.RemoteAddr()

// Get local port
port := t.LocalPort()

// Get last error
if err := t.LastError(); err != nil {
    log.Printf("Last error: %v", err)
}

// Update SSH config (requires restart)
t.UpdateConfig(newCfg)
t.Restart()
```

## Complete Example

```go
package main

import (
    "database/sql"
    "log"

    "github.com/pperesbr/gokit/pkg/tunnel"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    // Setup SSH config
    cfg, err := tunnel.NewSSHConfig(
        "ubuntu",
        "",
        "/home/user/.ssh/id_rsa",
        "bastion.example.com",
        "/home/user/.ssh/known_hosts",
        22,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create tunnel to internal database
    t := tunnel.NewTunnel(cfg, "db.internal", 5432, 0)
    
    if err := t.Start(); err != nil {
        log.Fatal(err)
    }
    defer t.Close()

    log.Printf("Tunnel established: %s -> %s", t.LocalAddr(), t.RemoteAddr())

    // Connect to database through tunnel
    connStr := fmt.Sprintf("postgres://user:pass@%s/mydb", t.LocalAddr())
    db, err := sql.Open("pgx", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Use database...
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    log.Println("Database connection successful!")
}
```

## Error Handling

```go
cfg, err := tunnel.NewSSHConfig(...)
if err != nil {
    // Possible errors:
    // - "host is required"
    // - "user is required"
    // - "password or keyFile is required"
    // - "failed to read keyFile: ..."
    // - "failed to parse keyFile: ..."
    // - "failed to load known_hosts: ..."
    log.Fatal(err)
}

t := tunnel.NewTunnel(cfg, "remote", 5432, 5432)
if err := t.Start(); err != nil {
    // Possible errors:
    // - "tunnel is already running"
    // - "config is required"
    // - "remoteHost is required"
    // - "remotePort must be greater than 0"
    // - "failed to connect to ssh server: ..."
    // - "failed to create local listener: ..."
    log.Fatal(err)
}
```

## Thread Safety

All tunnel operations are thread-safe. You can safely call `Status()`, `Stats()`, `LocalPort()`, etc. from multiple goroutines while the tunnel is running.