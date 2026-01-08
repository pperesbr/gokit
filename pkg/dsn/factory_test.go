package dsn_test

import (
	"testing"

	"github.com/pperesbr/gokit/pkg/dsn"
	"github.com/pperesbr/gokit/pkg/dsn/mysql"
	"github.com/pperesbr/gokit/pkg/dsn/oracle"
	"github.com/pperesbr/gokit/pkg/dsn/postgres"
)

func setupFactory() *dsn.Factory {
	f := dsn.NewFactory()
	f.Register("oracle", oracle.NewBuilder)
	f.Register("postgres", postgres.NewBuilder)
	f.Register("mysql", mysql.NewBuilder)
	return f
}

func TestFactory_LoadFromBytes_OracleStandalone(t *testing.T) {
	yaml := `
mode: standalone
host: db-server
port: 1521
service_name: ORCL
user: app
password: secret
`
	f := setupFactory()
	builder, err := f.BuildFromDriver("oracle", []byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder.Driver() != "oracle" {
		t.Errorf("Driver() = %q, want %q", builder.Driver(), "oracle")
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("ConnectionString() error: %v", err)
	}

	want := "app/secret@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=db-server)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=ORCL)))"
	if connStr != want {
		t.Errorf("ConnectionString() = %q, want %q", connStr, want)
	}
}

func TestFactory_LoadFromBytes_OracleRAC(t *testing.T) {
	yaml := `
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
`
	f := setupFactory()
	builder, err := f.BuildFromDriver("oracle", []byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder.Driver() != "oracle" {
		t.Errorf("Driver() = %q, want %q", builder.Driver(), "oracle")
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("ConnectionString() error: %v", err)
	}

	want := "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node1)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node2)(PORT=1521))(LOAD_BALANCE=ON)(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)))"
	if connStr != want {
		t.Errorf("ConnectionString() = %q, want %q", connStr, want)
	}
}

func TestFactory_LoadFromBytes_OracleDataGuard(t *testing.T) {
	yaml := `
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
failover_mode: SESSION
failover_retries: 30
failover_delay: 5
`
	f := setupFactory()
	builder, err := f.BuildFromDriver("oracle", []byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder.Driver() != "oracle" {
		t.Errorf("Driver() = %q, want %q", builder.Driver(), "oracle")
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("ConnectionString() error: %v", err)
	}

	want := "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=primary-db)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=standby-db1)(PORT=1521))(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)(FAILOVER_MODE=(TYPE=SESSION)(RETRIES=30)(DELAY=5))))"
	if connStr != want {
		t.Errorf("ConnectionString() = %q, want %q", connStr, want)
	}
}

func TestFactory_LoadFromBytes_OracleAutoDetect(t *testing.T) {
	yaml := `
oracle:
  mode: standalone
  host: db-server
  port: 1521
  service_name: ORCL
  user: app
  password: secret
`
	f := setupFactory()
	builder, err := f.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder.Driver() != "oracle" {
		t.Errorf("Driver() = %q, want %q", builder.Driver(), "oracle")
	}
}

func TestFactory_LoadFromBytes_Postgres(t *testing.T) {
	yaml := `
host: localhost
port: 5432
database: mydb
user: app
password: secret
`
	f := setupFactory()
	builder, err := f.BuildFromDriver("postgres", []byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder.Driver() != "postgres" {
		t.Errorf("Driver() = %q, want %q", builder.Driver(), "postgres")
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("ConnectionString() error: %v", err)
	}

	want := "postgres://app:secret@localhost:5432/mydb?sslmode=disable"
	if connStr != want {
		t.Errorf("ConnectionString() = %q, want %q", connStr, want)
	}
}

func TestFactory_LoadFromBytes_PostgresWithSSL(t *testing.T) {
	yaml := `
host: secure-db
port: 5432
database: mydb
user: app
password: secret
sslmode: require
connect_timeout: 10
`
	f := setupFactory()
	builder, err := f.BuildFromDriver("postgres", []byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("ConnectionString() error: %v", err)
	}

	want := "postgres://app:secret@secure-db:5432/mydb?connect_timeout=10&sslmode=require"
	if connStr != want {
		t.Errorf("ConnectionString() = %q, want %q", connStr, want)
	}
}

func TestFactory_LoadFromBytes_PostgresAutoDetect(t *testing.T) {
	yaml := `
postgres:
  host: localhost
  port: 5432
  database: mydb
  user: app
  password: secret
`
	f := setupFactory()
	builder, err := f.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder.Driver() != "postgres" {
		t.Errorf("Driver() = %q, want %q", builder.Driver(), "postgres")
	}
}

func TestFactory_LoadFromBytes_MySQL(t *testing.T) {
	yaml := `
host: localhost
port: 3306
database: mydb
user: app
password: secret
`
	f := setupFactory()
	builder, err := f.BuildFromDriver("mysql", []byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder.Driver() != "mysql" {
		t.Errorf("Driver() = %q, want %q", builder.Driver(), "mysql")
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("ConnectionString() error: %v", err)
	}

	want := "app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4"
	if connStr != want {
		t.Errorf("ConnectionString() = %q, want %q", connStr, want)
	}
}

func TestFactory_LoadFromBytes_MySQLWithOptions(t *testing.T) {
	yaml := `
host: db-server
port: 3306
database: mydb
user: app
password: secret
parse_time: true
charset: utf8mb4
timeout: 10
`
	f := setupFactory()
	builder, err := f.BuildFromDriver("mysql", []byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("ConnectionString() error: %v", err)
	}

	want := "app:secret@tcp(db-server:3306)/mydb?charset=utf8mb4&parseTime=true&timeout=10s"
	if connStr != want {
		t.Errorf("ConnectionString() = %q, want %q", connStr, want)
	}
}

func TestFactory_LoadFromBytes_MySQLAutoDetect(t *testing.T) {
	yaml := `
mysql:
  host: localhost
  port: 3306
  database: mydb
  user: app
  password: secret
`
	f := setupFactory()
	builder, err := f.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder.Driver() != "mysql" {
		t.Errorf("Driver() = %q, want %q", builder.Driver(), "mysql")
	}
}

func TestFactory_BuildFromDriver_UnknownDriver(t *testing.T) {
	f := setupFactory()
	_, err := f.BuildFromDriver("unknown", []byte("host: localhost"))
	if err == nil {
		t.Error("expected error for unknown driver, got nil")
	}
}

func TestFactory_LoadFromBytes_NoSupportedDriver(t *testing.T) {
	yaml := `
unknown:
  host: localhost
`
	f := setupFactory()
	_, err := f.LoadFromBytes([]byte(yaml))
	if err == nil {
		t.Error("expected error for unsupported driver, got nil")
	}
}

func TestFactory_LoadFromBytes_InvalidYAML(t *testing.T) {
	yaml := `
invalid: [yaml
`
	f := setupFactory()
	_, err := f.LoadFromBytes([]byte(yaml))
	if err == nil {
		t.Error("expected error for invalid yaml, got nil")
	}
}

func TestFactory_BuildFromDriver_InvalidConfig(t *testing.T) {
	yaml := `
mode: standalone
host: ""
user: ""
`
	f := setupFactory()
	builder, err := f.BuildFromDriver("oracle", []byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error creating builder: %v", err)
	}

	// Erro deve vir na validação, não na criação
	_, err = builder.ConnectionString()
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestFactory_Register(t *testing.T) {
	f := dsn.NewFactory()

	// Antes de registrar
	_, err := f.BuildFromDriver("oracle", []byte("mode: standalone"))
	if err == nil {
		t.Error("expected error before register, got nil")
	}

	// Depois de registrar
	f.Register("oracle", oracle.NewBuilder)
	_, err = f.BuildFromDriver("oracle", []byte(`
mode: standalone
host: localhost
service_name: ORCL
user: app
password: secret
`))
	if err != nil {
		t.Errorf("unexpected error after register: %v", err)
	}
}
