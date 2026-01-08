package dsn_test

import (
	"testing"

	"github.com/pperesbr/gokit/pkg/dsn"
	"github.com/pperesbr/gokit/pkg/dsn/oracle"
)

func setupFactory() *dsn.Factory {
	f := dsn.NewFactory()
	f.Register("oracle", oracle.NewBuilder)
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

func TestFactory_LoadFromBytes_AutoDetectDriver(t *testing.T) {
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

func TestFactory_BuildFromDriver_UnknownDriver(t *testing.T) {
	f := setupFactory()
	_, err := f.BuildFromDriver("mysql", []byte("host: localhost"))
	if err == nil {
		t.Error("expected error for unknown driver, got nil")
	}
}

func TestFactory_LoadFromBytes_NoSupportedDriver(t *testing.T) {
	yaml := `
mysql:
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
