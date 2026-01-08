package oracle

import (
	"testing"
)

func TestRACConfig_Validate(t *testing.T) {
	tests := []struct {
		name     string
		config   RACConfig
		wantErr  bool
		errField string
	}{
		{
			name: "valid with two nodes",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
					{Host: "rac-node2", Port: 1521},
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
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1"},
					{Host: "rac-node2"},
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
			name: "missing nodes",
			config: RACConfig{
				Nodes:       []Node{},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "nodes",
		},
		{
			name: "node missing host",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
					{Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "nodes[1].host",
		},
		{
			name: "node invalid port",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 70000},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "nodes[0].port",
		},
		{
			name: "missing service_name",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
				},
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
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
				},
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
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User: "app",
				},
			},
			wantErr:  true,
			errField: "password",
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

func TestRACConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name    string
		config  RACConfig
		want    string
		wantErr bool
	}{
		{
			name: "basic two nodes",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
					{Host: "rac-node2", Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "oracle://app:secret@rac-node1:1521,rac-node2:1521/ORCL",
		},
		{
			name: "with load balance and failover",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
					{Host: "rac-node2", Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				LoadBalance: true,
				Failover:    true,
			},
			want: "oracle://app:secret@rac-node1:1521,rac-node2:1521/ORCL?FAILOVER=true&LOAD_BALANCE=true",
		},
		{
			name: "with retry options",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				RetryCount: 3,
				RetryDelay: 5,
			},
			want: "oracle://app:secret@rac-node1:1521/ORCL?RETRY_COUNT=3&RETRY_DELAY=5",
		},
		{
			name: "with timeout",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 1521},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				Timeouts: Timeouts{
					ConnectTimeout: 10,
				},
			},
			want: "oracle://app:secret@rac-node1:1521/ORCL?TIMEOUT=10",
		},
		{
			name: "with default port",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1"},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "oracle://app:secret@rac-node1:1521/ORCL",
		},
		{
			name: "invalid config returns error",
			config: RACConfig{
				Nodes: []Node{},
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

func TestRACConfig_Driver(t *testing.T) {
	cfg := RACConfig{}

	if got := cfg.Driver(); got != DriverName {
		t.Errorf("Driver() = %q, want %q", got, DriverName)
	}
}
