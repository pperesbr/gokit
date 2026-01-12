package oracle

import (
	"errors"
	"testing"
)

func pint(i int) *int {
	return &i
}

func TestStandaloneConfig_Build(t *testing.T) {
	tests := []struct {
		name      string
		config    StandaloneConfig
		wantError error
		wantDSN   string
	}{
		{
			name: "valid config with no extra params",
			config: StandaloneConfig{
				Host:        "localhost",
				User:        "user",
				Password:    "password",
				ServiceName: "myservice",
				Port:        1521,
			},
			wantDSN: "oracle://user:password@localhost:1521/myservice",
		},
		{
			name: "valid config with extra params",
			config: StandaloneConfig{
				Host:              "localhost",
				User:              "user",
				Password:          "password",
				ServiceName:       "myservice",
				ConnectionTimeout: pint(10),
				Timeout:           pint(10),
			},
			wantDSN: "oracle://user:password@localhost:1521/myservice?CONNECTION TIMEOUT=10&TIMEOUT=10",
		},
		{
			name: "missing host",
			config: StandaloneConfig{
				User:        "user",
				Password:    "password",
				Port:        1521,
				ServiceName: "myservice",
			},
			wantError: ErrOracleHostRequired,
		},
		{
			name: "missing service name",
			config: StandaloneConfig{
				Host:     "localhost",
				User:     "user",
				Password: "password",
				Port:     1521,
			},
			wantError: ErrOracleServiceNameRequired,
		},
		{
			name: "missing user",
			config: StandaloneConfig{
				Host:        "localhost",
				ServiceName: "myservice",
				Password:    "password",
			},
			wantError: ErrOracleUserRequired,
		},
		{
			name: "missing password",
			config: StandaloneConfig{
				Host:        "localhost",
				ServiceName: "myservice",
				User:        "user",
			},
			wantError: ErrOraclePasswordRequired,
		},
		{
			name: "database port invalid (negative)",
			config: StandaloneConfig{
				Host:        "localhost",
				User:        "user",
				Password:    "password",
				Port:        -1521,
				ServiceName: "myservice",
			},
			wantError: ErrOraclePortInvalid,
		},
		{
			name: "database port invalid (too high)",
			config: StandaloneConfig{
				Host:        "localhost",
				User:        "user",
				Password:    "password",
				Port:        65536,
				ServiceName: "myservice",
			},
			wantError: ErrOraclePortInvalid,
		},
		{
			name: "config with default port",
			config: StandaloneConfig{
				Host:        "localhost",
				User:        "user",
				Password:    "password",
				ServiceName: "myservice",
			},
			wantDSN: "oracle://user:password@localhost:1521/myservice",
		},
		{
			name: "extra param: connection_timeout with negative value",
			config: StandaloneConfig{
				Host:              "localhost",
				User:              "user",
				Password:          "password",
				ServiceName:       "myservice",
				ConnectionTimeout: pint(-1),
			},
			wantError: ErrOracleConnectTimeoutInvalid,
		},
		{
			name: "extra param: connection_timeout greater than or equal to 0",
			config: StandaloneConfig{
				Host:              "localhost",
				User:              "user",
				Password:          "password",
				ServiceName:       "myservice",
				ConnectionTimeout: pint(10),
			},
			wantDSN: "oracle://user:password@localhost:1521/myservice?CONNECTION TIMEOUT=10",
		},
		{
			name: "extra param: timeout with negative value",
			config: StandaloneConfig{
				Host:        "localhost",
				User:        "user",
				Password:    "password",
				ServiceName: "myservice",
				Timeout:     pint(-1),
			},
			wantError: ErrOracleTimeoutInvalid,
		},
		{
			name: "extra param: timeout greater than or equal to 0",
			config: StandaloneConfig{
				Host:        "localhost",
				User:        "user",
				Password:    "password",
				ServiceName: "myservice",
				Timeout:     pint(10),
			},
			wantDSN: "oracle://user:password@localhost:1521/myservice?TIMEOUT=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, err := tt.config.Build()

			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Errorf("error: got %v, want %v", err, tt.wantError)
					return
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if dsn != tt.wantDSN {
				t.Errorf("dsn: got %s, want %s", dsn, tt.wantDSN)
			}
		})
	}
}
