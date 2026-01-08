package postgres

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with sslmode",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				SSLMode:  "require",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: Config{
				Port:     5432,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "postgres: host is required",
		},
		{
			name: "missing database",
			config: Config{
				Host: "localhost",
				Port: 5432,
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "postgres: database is required",
		},
		{
			name: "missing user",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Credentials: Credentials{
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "postgres: user is required",
		},
		{
			name: "missing password",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Credentials: Credentials{
					User: "app",
				},
			},
			wantErr: true,
			errMsg:  "postgres: password is required",
		},
		{
			name: "invalid port - negative",
			config: Config{
				Host:     "localhost",
				Port:     -1,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "postgres: port must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			config: Config{
				Host:     "localhost",
				Port:     70000,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "postgres: port must be between 1 and 65535",
		},
		{
			name: "invalid sslmode",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				SSLMode:  "invalid",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "postgres: sslmode must be one of: disable, require, verify-ca, verify-full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		want    string
		wantErr bool
	}{
		{
			name: "basic connection string",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want:    "postgres://app:secret@localhost:5432/mydb?sslmode=disable",
			wantErr: false,
		},
		{
			name: "with default port",
			config: Config{
				Host:     "db-server",
				Database: "production",
				Credentials: Credentials{
					User:     "admin",
					Password: "pass123",
				},
			},
			want:    "postgres://admin:pass123@db-server:5432/production?sslmode=disable",
			wantErr: false,
		},
		{
			name: "with sslmode require",
			config: Config{
				Host:     "secure-db",
				Port:     5432,
				Database: "mydb",
				SSLMode:  "require",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want:    "postgres://app:secret@secure-db:5432/mydb?sslmode=require",
			wantErr: false,
		},
		{
			name: "with connect timeout",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				Timeouts: Timeouts{
					ConnectTimeout: 10,
				},
			},
			want:    "postgres://app:secret@localhost:5432/mydb?connect_timeout=10&sslmode=disable",
			wantErr: false,
		},
		{
			name: "with special characters in password",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "p@ss:word/123",
				},
			},
			want:    "postgres://app:p%40ss%3Aword%2F123@localhost:5432/mydb?sslmode=disable",
			wantErr: false,
		},
		{
			name: "full config",
			config: Config{
				Host:     "prod-db.example.com",
				Port:     5433,
				Database: "analytics",
				SSLMode:  "verify-full",
				Credentials: Credentials{
					User:     "analyst",
					Password: "secure123",
				},
				Timeouts: Timeouts{
					ConnectTimeout: 30,
				},
			},
			want:    "postgres://analyst:secure123@prod-db.example.com:5433/analytics?connect_timeout=30&sslmode=verify-full",
			wantErr: false,
		},
		{
			name: "invalid config returns error",
			config: Config{
				Port:     5432,
				Database: "mydb",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.ConnectionString()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConnectionString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConnectionString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Driver(t *testing.T) {
	c := &Config{}
	if got := c.Driver(); got != DriverName {
		t.Errorf("Driver() = %v, want %v", got, DriverName)
	}
}

func TestIsValidSSLMode(t *testing.T) {
	tests := []struct {
		mode  string
		valid bool
	}{
		{"disable", true},
		{"require", true},
		{"verify-ca", true},
		{"verify-full", true},
		{"", false},
		{"invalid", false},
		{"DISABLE", false},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			if got := isValidSSLMode(tt.mode); got != tt.valid {
				t.Errorf("isValidSSLMode(%q) = %v, want %v", tt.mode, got, tt.valid)
			}
		})
	}
}

func TestNewBuilder(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		wantDSN string
	}{
		{
			name: "valid yaml",
			yaml: `
host: localhost
port: 5432
database: mydb
user: app
password: secret
`,
			wantErr: false,
			wantDSN: "postgres://app:secret@localhost:5432/mydb?sslmode=disable",
		},
		{
			name: "valid yaml with sslmode",
			yaml: `
host: localhost
port: 5432
database: mydb
user: app
password: secret
sslmode: require
`,
			wantErr: false,
			wantDSN: "postgres://app:secret@localhost:5432/mydb?sslmode=require",
		},
		{
			name: "valid yaml with timeout",
			yaml: `
host: localhost
port: 5432
database: mydb
user: app
password: secret
connect_timeout: 10
`,
			wantErr: false,
			wantDSN: "postgres://app:secret@localhost:5432/mydb?connect_timeout=10&sslmode=disable",
		},
		{
			name: "full yaml config",
			yaml: `
host: prod-db.example.com
port: 5433
database: analytics
sslmode: verify-full
user: analyst
password: secure123
connect_timeout: 30
`,
			wantErr: false,
			wantDSN: "postgres://analyst:secure123@prod-db.example.com:5433/analytics?connect_timeout=30&sslmode=verify-full",
		},
		{
			name:    "invalid yaml",
			yaml:    `invalid: [yaml`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewBuilder([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if builder.Driver() != DriverName {
				t.Errorf("Driver() = %v, want %v", builder.Driver(), DriverName)
			}

			dsn, err := builder.ConnectionString()
			if err != nil {
				t.Errorf("ConnectionString() error = %v", err)
				return
			}

			if dsn != tt.wantDSN {
				t.Errorf("ConnectionString() = %v, want %v", dsn, tt.wantDSN)
			}
		})
	}
}
