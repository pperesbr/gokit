package oracle

import (
	"testing"
)

func TestNewBuilder_InvalidMode(t *testing.T) {
	yaml := `
mode: invalid
host: localhost
service_name: ORCL
user: app
password: secret
`
	_, err := NewBuilder([]byte(yaml))
	if err == nil {
		t.Error("expected error for invalid mode, got nil")
	}
}

func TestNewBuilder_InvalidYAML(t *testing.T) {
	yaml := `invalid: [yaml`
	_, err := NewBuilder([]byte(yaml))
	if err == nil {
		t.Error("expected error for invalid yaml, got nil")
	}
}

func TestNewBuilder_DefaultMode(t *testing.T) {
	yaml := `
host: localhost
service_name: ORCL
user: app
password: secret
`
	builder, err := NewBuilder([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := builder.(*StandaloneConfig)
	if !ok {
		t.Error("expected StandaloneConfig when mode is empty")
	}
}

func TestNewBuilder_Standalone(t *testing.T) {
	yaml := `
mode: standalone
host: db-server
port: 1521
service_name: ORCL
user: app
password: secret
connect_timeout: 10
`
	builder, err := NewBuilder([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, ok := builder.(*StandaloneConfig)
	if !ok {
		t.Fatal("expected StandaloneConfig")
	}

	if cfg.Host != "db-server" {
		t.Errorf("Host = %q, want %q", cfg.Host, "db-server")
	}

	if cfg.Port != 1521 {
		t.Errorf("Port = %d, want %d", cfg.Port, 1521)
	}

	if cfg.ServiceName != "ORCL" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "ORCL")
	}

	if cfg.User != "app" {
		t.Errorf("User = %q, want %q", cfg.User, "app")
	}

	if cfg.ConnectTimeout != 10 {
		t.Errorf("ConnectTimeout = %d, want %d", cfg.ConnectTimeout, 10)
	}
}

func TestNewBuilder_RAC(t *testing.T) {
	yaml := `
mode: rac
service_name: ORCL
user: app
password: secret
nodes:
  - host: rac-node1
    port: 1521
  - host: rac-node2
    port: 1522
    protocol: TCPS
load_balance: true
failover: true
retry_count: 3
`
	builder, err := NewBuilder([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, ok := builder.(*RACConfig)
	if !ok {
		t.Fatal("expected RACConfig")
	}

	if len(cfg.Nodes) != 2 {
		t.Fatalf("Nodes length = %d, want %d", len(cfg.Nodes), 2)
	}

	if cfg.Nodes[0].Host != "rac-node1" {
		t.Errorf("Nodes[0].Host = %q, want %q", cfg.Nodes[0].Host, "rac-node1")
	}

	if cfg.Nodes[1].Port != 1522 {
		t.Errorf("Nodes[1].Port = %d, want %d", cfg.Nodes[1].Port, 1522)
	}

	if cfg.Nodes[1].Protocol != "TCPS" {
		t.Errorf("Nodes[1].Protocol = %q, want %q", cfg.Nodes[1].Protocol, "TCPS")
	}

	if !cfg.LoadBalance {
		t.Error("LoadBalance = false, want true")
	}

	if !cfg.Failover {
		t.Error("Failover = false, want true")
	}

	if cfg.RetryCount != 3 {
		t.Errorf("RetryCount = %d, want %d", cfg.RetryCount, 3)
	}
}

func TestNewBuilder_DataGuard(t *testing.T) {
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
  - host: standby-db2
    port: 1521
failover_mode: SESSION
failover_retries: 30
failover_delay: 5
`
	builder, err := NewBuilder([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, ok := builder.(*DataGuardConfig)
	if !ok {
		t.Fatal("expected DataGuardConfig")
	}

	if cfg.Primary.Host != "primary-db" {
		t.Errorf("Primary.Host = %q, want %q", cfg.Primary.Host, "primary-db")
	}

	if len(cfg.Standbys) != 2 {
		t.Fatalf("Standbys length = %d, want %d", len(cfg.Standbys), 2)
	}

	if cfg.Standbys[0].Host != "standby-db1" {
		t.Errorf("Standbys[0].Host = %q, want %q", cfg.Standbys[0].Host, "standby-db1")
	}

	if cfg.FailoverMode != "SESSION" {
		t.Errorf("FailoverMode = %q, want %q", cfg.FailoverMode, "SESSION")
	}

	if cfg.FailoverRetries != 30 {
		t.Errorf("FailoverRetries = %d, want %d", cfg.FailoverRetries, 30)
	}

	if cfg.FailoverDelay != 5 {
		t.Errorf("FailoverDelay = %d, want %d", cfg.FailoverDelay, 5)
	}
}
