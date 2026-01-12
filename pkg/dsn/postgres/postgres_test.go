package postgres

import (
	"errors"
	"testing"
)

func pint(v int) *int {
	return &v
}

func TestIsValidSSLMode(t *testing.T) {
	tests := []struct {
		mode string
		want bool
	}{
		{"disable", true},
		{"allow", true},
		{"prefer", true},
		{"require", true},
		{"verify-ca", true},
		{"verify-full", true},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			if got := isValidSSLMode(tt.mode); got != tt.want {
				t.Errorf("isValidSSLMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Build(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
		wantDSN string
	}{
		{
			name: "valid config with no extra params",
			config: Config{
				Host:     "localhost",
				User:     "user",
				Password: "password",
				Database: "mydb",
				Port:     5432,
			},
			wantDSN: "postgres://user:password@localhost:5432/mydb",
		},
		{
			name: "valid config with extra params",
			config: Config{
				Host:            "localhost",
				User:            "user",
				Password:        "password",
				Database:        "mydb",
				Port:            5432,
				SSLMode:         "verify-full",
				ApplicationName: "myapp",
				ConnectTimeout:  pint(0),
				SearchPath:      "myapp,public",
				Timezone:        "America/Sao_Paulo",
			},
			wantDSN: "postgres://user:password@localhost:5432/mydb?sslmode=verify-full&application_name=myapp&connect_timeout=0&search_path=myapp%2Cpublic&timezone=America%2FSao_Paulo",
		},
		{
			name: "missing host field",
			config: Config{
				Database: "mydb",
				User:     "user",
				Password: "password",
				Port:     5432,
			},
			wantErr: ErrPostgresHostRequired,
		},
		{
			name: "missing user field",
			config: Config{
				Host:     "localhost",
				Database: "mydb",
				Password: "password",
				Port:     5432,
			},
			wantErr: ErrPostgresUserRequired,
		},
		{
			name: "missing password field",
			config: Config{
				Host:     "localhost",
				Database: "mydb",
				User:     "user",
				Port:     5432,
			},
			wantErr: ErrPostgresPasswordRequired,
		},
		{
			name: "missing database field",
			config: Config{
				Host:     "localhost",
				User:     "user",
				Password: "password",
				Port:     5432,
			},
			wantErr: ErrPostgresDatabaseRequired,
		},
		{
			name: "database port invalid (negative)",
			config: Config{
				Host:     "localhost",
				User:     "user",
				Password: "password",
				Database: "mydb",
				Port:     -5432,
			},
			wantErr: ErrPostgresInvalidPort,
		},
		{
			name: "database port invalid (too high)",
			config: Config{
				Host:     "localhost",
				User:     "user",
				Password: "password",
				Database: "mydb",
				Port:     65536,
			},
			wantErr: ErrPostgresInvalidPort,
		},
		{
			name: "set default port when value is zero",
			config: Config{
				Host:     "localhost",
				User:     "user",
				Password: "password",
				Database: "mydb",
			},
			wantDSN: "postgres://user:password@localhost:5432/mydb",
		}, {
			name: "extra param: sslmode invalid",
			config: Config{
				Host:     "localhost",
				User:     "user",
				Password: "password",
				Database: "mydb",
				Port:     5432,
				SSLMode:  "sslmode invalid",
			},
			wantErr: ErrPostgresInvalidSSLMode,
		}, {
			name: "extra param: connect_timeout with negative value",
			config: Config{
				Host:           "localhost",
				User:           "user",
				Password:       "password",
				Database:       "mydb",
				Port:           5432,
				ConnectTimeout: pint(-1),
			},
			wantErr: ErrPostgresInvalidConnectTimeout,
		},
		{
			name: "extra param: connect_timeout with zero value (disable connect_timeout)",
			config: Config{
				Host:           "localhost",
				User:           "user",
				Password:       "password",
				Database:       "mydb",
				Port:           5432,
				ConnectTimeout: pint(0),
			},
			wantDSN: "postgres://user:password@localhost:5432/mydb?connect_timeout=0",
		},
		{
			name: "special characters in password",
			config: Config{
				Host:     "localhost",
				User:     "user",
				Password: "p@ss:word/special",
				Database: "mydb",
				Port:     5432,
			},
			wantDSN: "postgres://user:p%40ss%3Aword%2Fspecial@localhost:5432/mydb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, err := tt.config.Build()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error: got %v, want %v", err, tt.wantErr)
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
