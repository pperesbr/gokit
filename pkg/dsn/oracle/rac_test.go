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
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node1)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node2)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
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
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node1)(PORT=1521))(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node2)(PORT=1521))(LOAD_BALANCE=ON)(FAILOVER=ON))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
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
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node1)(PORT=1521))(RETRY_COUNT=3)(RETRY_DELAY=5))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
		},
		{
			name: "with timeouts",
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
					ConnectTimeout:          10,
					TransportConnectTimeout: 5,
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node1)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=ORCL))(CONNECT_TIMEOUT=10)(TRANSPORT_CONNECT_TIMEOUT=5))",
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
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCP)(HOST=rac-node1)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
		},
		{
			name: "with tcps protocol",
			config: RACConfig{
				Nodes: []Node{
					{Host: "rac-node1", Port: 2484, Protocol: "TCPS"},
				},
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=TCPS)(HOST=rac-node1)(PORT=2484)))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
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
