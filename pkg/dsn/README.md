# DSN - Database Connection String Builder

Package for generating connection strings for different databases. Supports Oracle (Standalone, RAC, DataGuard), PostgreSQL and MySQL.

## Installation

```go
import "github.com/pperesbr/gokit/pkg/dsn"
```

## Supported Drivers

| Driver     | Modes                        | Recommended Go Driver          |
|------------|------------------------------|--------------------------------|
| Oracle     | Standalone, RAC, DataGuard   | github.com/sijms/go-ora/v2     |
| PostgreSQL | -                            | github.com/jackc/pgx/v5/stdlib |
| MySQL      | -                            | github.com/go-sql-driver/mysql |

## Direct Usage

### Oracle Standalone

```go
import "github.com/pperesbr/gokit/pkg/dsn/oracle"

cfg := &oracle.StandaloneConfig{
    Host:        "db-server",
    Port:        1521,
    ServiceName: "ORCL",
    Credentials: oracle.Credentials{
        User:     "app",
        Password: "secret",
    },
}

connStr, err := cfg.ConnectionString()
// oracle://app:secret@db-server:1521/ORCL
```

### Oracle RAC

```go
cfg := &oracle.RACConfig{
    Nodes: []oracle.Node{
        {Host: "rac-node1", Port: 1521},
        {Host: "rac-node2", Port: 1521},
    },
    ServiceName: "ORCL",
    Credentials: oracle.Credentials{
        User:     "app",
        Password: "secret",
    },
    LoadBalance: true,
    Failover:    true,
}

connStr, err := cfg.ConnectionString()
// oracle://app:secret@rac-node1:1521,rac-node2:1521/ORCL?FAILOVER=true&LOAD_BALANCE=true
```

### Oracle DataGuard

```go
cfg := &oracle.DataGuardConfig{
    Primary:  oracle.Node{Host: "primary-db", Port: 1521},
    Standbys: []oracle.Node{{Host: "standby-db1", Port: 1521}},
    ServiceName: "ORCL",
    Credentials: oracle.Credentials{
        User:     "app",
        Password: "secret",
    },
    FailoverMode:    oracle.FailoverModeSession,
    FailoverRetries: 30,
    FailoverDelay:   5,
}

connStr, err := cfg.ConnectionString()
// oracle://app:secret@primary-db:1521,standby-db1:1521/ORCL?FAILOVER=true&FAILOVER_DELAY=5&FAILOVER_MODE=SESSION&FAILOVER_RETRIES=30
```

### PostgreSQL

```go
import "github.com/pperesbr/gokit/pkg/dsn/postgres"

cfg := &postgres.Config{
    Host:     "localhost",
    Port:     5432,
    Database: "mydb",
    SSLMode:  "disable",
    Credentials: postgres.Credentials{
        User:     "app",
        Password: "secret",
    },
}

connStr, err := cfg.ConnectionString()
// postgres://app:secret@localhost:5432/mydb?sslmode=disable
```

### MySQL

```go
import "github.com/pperesbr/gokit/pkg/dsn/mysql"

cfg := &mysql.Config{
    Host:      "localhost",
    Port:      3306,
    Database:  "mydb",
    ParseTime: true,
    Credentials: mysql.Credentials{
        User:     "app",
        Password: "secret",
    },
}

connStr, err := cfg.ConnectionString()
// app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true
```

## Factory Usage

The Factory allows loading configurations from YAML files and automatically detecting the driver.

### Register Builders

```go
import (
    "github.com/pperesbr/gokit/pkg/dsn"
    "github.com/pperesbr/gokit/pkg/dsn/oracle"
    "github.com/pperesbr/gokit/pkg/dsn/postgres"
    "github.com/pperesbr/gokit/pkg/dsn/mysql"
)

factory := dsn.NewFactory()
factory.Register("oracle", oracle.NewBuilder)
factory.Register("postgres", postgres.NewBuilder)
factory.Register("mysql", mysql.NewBuilder)
```

### Load from YAML (auto-detect)

```yaml
# config.yaml
postgres:
  host: localhost
  port: 5432
  database: mydb
  user: app
  password: secret
  sslmode: disable
```

```go
builder, err := factory.LoadFromYAML("config.yaml")
if err != nil {
    log.Fatal(err)
}

connStr, err := builder.ConnectionString()
```

### Load by Specific Driver

```yaml
# oracle.yaml
mode: standalone
host: db-server
port: 1521
service_name: ORCL
user: app
password: secret
```

```go
data, _ := os.ReadFile("oracle.yaml")
builder, err := factory.BuildFromDriver("oracle", data)
```

## YAML Configuration by Driver

### Oracle Standalone

```yaml
mode: standalone
host: db-server
port: 1521
service_name: ORCL  # or sid: ORCL
user: app
password: secret
connect_timeout: 10
```

### Oracle RAC

```yaml
mode: rac
service_name: ORCL
user: app
password: secret
nodes:
  - host: rac-node1
    port: 1521
  - host: rac-node2
    port: 1521
load_balance: true
failover: true
retry_count: 3
retry_delay: 5
```

### Oracle DataGuard

```yaml
mode: dataguard
service_name: ORCL
user: app
password: secret
primary:
  host: primary-db
  port: 1521
standbys:
  - host: standby-db1
    port: 1521
failover_mode: SESSION  # or SELECT
failover_retries: 30
failover_delay: 5
```

### PostgreSQL

```yaml
host: localhost
port: 5432
database: mydb
user: app
password: secret
sslmode: disable  # disable, require, verify-ca, verify-full
connect_timeout: 10
```

### MySQL

```yaml
host: localhost
port: 3306
database: mydb
user: app
password: secret
charset: utf8mb4
parse_time: true
loc: Local
timeout: 10
read_timeout: 30
write_timeout: 30
```

## Validation

All configurations are validated before generating the connection string:

```go
cfg := &postgres.Config{
    Host: "localhost",
    // Database not provided
}

connStr, err := cfg.ConnectionString()
// err: "postgres: database is required"
```

## Integration Tests

The package includes integration tests using testcontainers. To run:

```bash
go test -tags=integration -v ./pkg/dsn/...
```

Requires Docker running.