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
			want: "oracle://app:secret@primary-db:1521,standby-db1:1521/ORCL?FAILOVER=true",
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
			want: "oracle://app:secret@primary-db:1521,standby-db1:1521,standby-db2:1521/ORCL?FAILOVER=true",
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
			want: "oracle://app:secret@primary-db:1521,standby-db1:1521/ORCL?FAILOVER=true&FAILOVER_MODE=SESSION",
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
			want: "oracle://app:secret@primary-db:1521,standby-db1:1521/ORCL?FAILOVER=true&FAILOVER_DELAY=5&FAILOVER_MODE=SESSION&FAILOVER_RETRIES=30",
		},
		{
			name: "with timeout",
			config: DataGuardConfig{
				Primary:     Node{Host: "primary-db", Port: 1521},
				Standbys:    []Node{{Host: "standby-db1", Port: 1521}},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				Timeouts: Timeouts{
					ConnectTimeout: 10,
				},
			},
			want: "oracle://app:secret@primary-db:1521,standby-db1:1521/ORCL?FAILOVER=true&TIMEOUT=10",
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
			want: "oracle://app:secret@primary-db:1521,standby-db1:1521/ORCL?FAILOVER=true",
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
