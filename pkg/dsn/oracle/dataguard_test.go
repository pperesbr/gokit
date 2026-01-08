package oracle

import (
	"testing"
)

func TestDataGuardConfig_Validate(t *testing.T) {
	tests := []struct {
		name     string
		config   DataGuardConfig
		wantErr  bool
		errField string
	}{
		{
			name: "valid with one standby",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "valid with multiple standbys",
			config: DataGuardConfig{
				Primary: Node{Host: "primary-db", Port: 1521},
				Standbys: []Node{
					{Host: "standby-db1", Port: 1521},
					{Host: "standby-db2", Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "valid with default port",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db"},
				Standbys:    []Node{{Host: "standby-db1"}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "missing primary host",
			config: DataGuardConfig{
				Primary:     Node{Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "primary.host",
		},
		{
			name: "invalid primary port",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 70000},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "primary.port",
		},
		{
			name: "missing standbys",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "standbys",
		},
		{
			name: "standby missing host",
			config: DataGuardConfig{
				Primary: Node{Host: "primary-db", Port: 1521},
				Standbys: []Node{
					{Host: "standby-db1", Port: 1521},
					{Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "standbys[1].host",
		},
		{
			name: "standby invalid port",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 70000}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "standbys[0].port",
		},
		{
			name: "missing service_name",
			config: DataGuardConfig{
				Primary:  Node{Host: "primary-db", Port: 1521},
				Standbys: []Node{{Host: "standby-db1", Port: 1521}},
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "service_name",
		},
		{
			name: "missing user",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "user",
		},
		{
			name: "missing password",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User: "app",
				},
			},
			wantErr:  true,
			errField: "password",
		},
		{
			name: "invalid failover_mode",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				FailoverMode: "INVALID",
			},
			wantErr:  true,
			errField: "failover_mode",
		},
		{
			name: "valid failover_mode SESSION",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				FailoverMode: FailoverModeSession,
			},
			wantErr: false,
		},
		{
			name: "valid failover_mode SELECT",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				FailoverMode: FailoverModeSelect,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDataGuardConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name    string
		config  DataGuardConfig
		want    string
		wantErr bool
	}{
		{
			name: "basic with one standby",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=primary-db)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=standby-db1)(PORT=1521))(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
		},
		{
			name: "with multiple standbys",
			config: DataGuardConfig{
				Primary: Node{Host: "primary-db", Port: 1521},
				Standbys: []Node{
					{Host: "standby-db1", Port: 1521},
					{Host: "standby-db2", Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=primary-db)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=standby-db1)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=standby-db2)(PORT=1521))(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
		},
		{
			name: "with failover mode SESSION",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				FailoverMode: FailoverModeSession,
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=primary-db)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=standby-db1)(PORT=1521))(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)(FAILOVER_MODE=(TYPE=SESSION))))",
		},
		{
			name: "with failover mode and retries",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				FailoverMode:    FailoverModeSession,
				FailoverRetries: 30,
				FailoverDelay:   5,
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=primary-db)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=standby-db1)(PORT=1521))(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)(FAILOVER_MODE=(TYPE=SESSION)(RETRIES=30)(DELAY=5))))",
		},
		{
			name: "with timeouts",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				Timeouts: Timeouts{
					ConnectTimeout:          10,
					TransportConnectTimeout: 5,
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=primary-db)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=standby-db1)(PORT=1521))(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL))(CONNECT_TIMEOUT=10)(TRANSPORT_CONNECT_TIMEOUT=5))",
		},
		{
			name: "with default port",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db"},
				Standbys:    []Node{{Host: "standby-db1"}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=primary-db)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=standby-db1)(PORT=1521))(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
		},
		{
			name: "with tcps protocol",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 2484, Protocol: "TCPS"},
				Standbys:    []Node{{Host: "standby-db1", Port: 2484, Protocol: "TCPS"}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCPS)(HOST=primary-db)(PORT=2484))(ADDRESS=(PROTOCOL=TCPS)(HOST=standby-db1)(PORT=2484))(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
		},
		{
			name: "invalid config returns error",
			config: DataGuardConfig{
				Primary: Node{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.ConnectionString()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("ConnectionString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDataGuardConfig_Driver(t *testing.T) {
	cfg := DataGuardConfig{}

	if got := cfg.Driver(); got != DriverName {
		t.Errorf("Driver() = %q, want %q", got, DriverName)
	}
}
