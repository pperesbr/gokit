package oracle

import (
	"testing"
)

func TestStandaloneConfig_Validate(t *testing.T) {
	tests := []struct {
		name     string
		config   StandaloneConfig
		wantErr  bool
		errField string
	}{
		{
			name: "valid with service_name",
			config: StandaloneConfig{
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "valid with sid",
			config: StandaloneConfig{
				Host: "localhost",
				Port: 1521,
				SID:  "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "valid without port uses default",
			config: StandaloneConfig{
				Host:        "localhost",
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: StandaloneConfig{
				Port:        1521,
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "host",
		},
		{
			name: "missing service_name and sid",
			config: StandaloneConfig{
				Host: "localhost",
				Port: 1521,
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "service_name/sid",
		},
		{
			name: "missing user",
			config: StandaloneConfig{
				Host:        "localhost",
				Port:        1521,
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
			config: StandaloneConfig{
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Credentials: Credentials{
					User: "app",
				},
			},
			wantErr:  true,
			errField: "password",
		},
		{
			name: "invalid port",
			config: StandaloneConfig{
				Host:        "localhost",
				Port:        70000,
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr:  true,
			errField: "port",
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

func TestStandaloneConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name    string
		config  StandaloneConfig
		want    string
		wantErr bool
	}{
		{
			name: "basic with service_name",
			config: StandaloneConfig{
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=localhost)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
		},
		{
			name: "with sid",
			config: StandaloneConfig{
				Host: "localhost",
				Port: 1521,
				SID:  "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=localhost)(PORT=1521))(CONNECT_DATA=(SID=ORCL)))",
		},
		{
			name: "default port",
			config: StandaloneConfig{
				Host:        "localhost",
				ServiceName: "ORCL",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want: "app/secret@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=localhost)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=ORCL)))",
		},
		{
			name: "with timeouts",
			config: StandaloneConfig{
				Host:        "localhost",
				Port:        1521,
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
			want: "app/secret@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=localhost)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=ORCL))(CONNECT_TIMEOUT=10)(TRANSPORT_CONNECT_TIMEOUT=5))",
		},
		{
			name: "invalid config returns error",
			config: StandaloneConfig{
				Port: 1521,
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

func TestStandaloneConfig_Driver(t *testing.T) {
	cfg := StandaloneConfig{}

	if got := cfg.Driver(); got != DriverName {
		t.Errorf("Driver() = %q, want %q", got, DriverName)
	}
}
