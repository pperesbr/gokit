package mysql

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
				Port:     3306,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with protocol",
			config: Config{
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Protocol: "tcp",
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
				Port:     3306,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "mysql: host is required",
		},
		{
			name: "missing database",
			config: Config{
				Host: "localhost",
				Port: 3306,
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "mysql: database is required",
		},
		{
			name: "missing user",
			config: Config{
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Credentials: Credentials{
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "mysql: user is required",
		},
		{
			name: "missing password",
			config: Config{
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Credentials: Credentials{
					User: "app",
				},
			},
			wantErr: true,
			errMsg:  "mysql: password is required",
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
			errMsg:  "mysql: port must be between 1 and 65535",
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
			errMsg:  "mysql: port must be between 1 and 65535",
		},
		{
			name: "invalid protocol",
			config: Config{
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Protocol: "invalid",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			wantErr: true,
			errMsg:  "mysql: protocol must be one of: tcp, unix",
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
				Port:     3306,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want:    "app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4",
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
			want:    "admin:pass123@tcp(db-server:3306)/production?charset=utf8mb4",
			wantErr: false,
		},
		{
			name: "with parseTime",
			config: Config{
				Host:      "localhost",
				Port:      3306,
				Database:  "mydb",
				ParseTime: true,
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want:    "app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true",
			wantErr: false,
		},
		{
			name: "with loc",
			config: Config{
				Host:      "localhost",
				Port:      3306,
				Database:  "mydb",
				ParseTime: true,
				Loc:       "Local",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want:    "app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true&loc=Local",
			wantErr: false,
		},
		{
			name: "with timeouts",
			config: Config{
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
				Timeouts: Timeouts{
					Timeout:      10,
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
			},
			want:    "app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4&timeout=10s&readTimeout=30s&writeTimeout=30s",
			wantErr: false,
		},
		{
			name: "with custom charset",
			config: Config{
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Charset:  "latin1",
				Credentials: Credentials{
					User:     "app",
					Password: "secret",
				},
			},
			want:    "app:secret@tcp(localhost:3306)/mydb?charset=latin1",
			wantErr: false,
		},
		{
			name: "with special characters in password",
			config: Config{
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Credentials: Credentials{
					User:     "app",
					Password: "p@ss:word/123",
				},
			},
			want:    "app:p%40ss%3Aword%2F123@tcp(localhost:3306)/mydb?charset=utf8mb4",
			wantErr: false,
		},
		{
			name: "full config",
			config: Config{
				Host:      "prod-db.example.com",
				Port:      3307,
				Database:  "analytics",
				Protocol:  "tcp",
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "UTC",
				Credentials: Credentials{
					User:     "analyst",
					Password: "secure123",
				},
				Timeouts: Timeouts{
					Timeout:      5,
					ReadTimeout:  10,
					WriteTimeout: 10,
				},
			},
			want:    "analyst:secure123@tcp(prod-db.example.com:3307)/analytics?charset=utf8mb4&parseTime=true&loc=UTC&timeout=5s&readTimeout=10s&writeTimeout=10s",
			wantErr: false,
		},
		{
			name: "invalid config returns error",
			config: Config{
				Port:     3306,
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

func TestIsValidProtocol(t *testing.T) {
	tests := []struct {
		protocol string
		valid    bool
	}{
		{"tcp", true},
		{"unix", true},
		{"", false},
		{"invalid", false},
		{"TCP", false},
	}

	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			if got := isValidProtocol(tt.protocol); got != tt.valid {
				t.Errorf("isValidProtocol(%q) = %v, want %v", tt.protocol, got, tt.valid)
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
port: 3306
database: mydb
user: app
password: secret
`,
			wantErr: false,
			wantDSN: "app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4",
		},
		{
			name: "valid yaml with parseTime",
			yaml: `
host: localhost
port: 3306
database: mydb
user: app
password: secret
parse_time: true
`,
			wantErr: false,
			wantDSN: "app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true",
		},
		{
			name: "valid yaml with timeouts",
			yaml: `
host: localhost
port: 3306
database: mydb
user: app
password: secret
timeout: 10
read_timeout: 30
write_timeout: 30
`,
			wantErr: false,
			wantDSN: "app:secret@tcp(localhost:3306)/mydb?charset=utf8mb4&timeout=10s&readTimeout=30s&writeTimeout=30s",
		},
		{
			name: "full yaml config",
			yaml: `
host: prod-db.example.com
port: 3307
database: analytics
protocol: tcp
charset: utf8mb4
parse_time: true
loc: UTC
user: analyst
password: secure123
timeout: 5
read_timeout: 10
write_timeout: 10
`,
			wantErr: false,
			wantDSN: "analyst:secure123@tcp(prod-db.example.com:3307)/analytics?charset=utf8mb4&parseTime=true&loc=UTC&timeout=5s&readTimeout=10s&writeTimeout=10s",
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
