package mysql

import (
	"errors"
	"testing"
)

func pint(i int) *int {
	return &i
}

func pbool(b bool) *bool {
	return &b
}

func TestConfig_Build(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError error
		wantDSN   string
	}{
		{
			name: "valid config with no extra params",
			config: Config{
				Host:     "localhost",
				User:     "root",
				Password: "secret",
				Database: "mydb",
				Port:     3306,
			},
			wantDSN: "root:secret@tcp(localhost:3306)/mydb",
		},
		{
			name: "valid config with extra params",
			config: Config{
				Host:         "localhost",
				User:         "root",
				Password:     "secret",
				Database:     "mydb",
				Port:         3306,
				Charset:      "utf8mb4",
				ParseTime:    pbool(true),
				Loc:          "Local",
				Timeout:      pint(5),
				ReadTimeout:  pint(30),
				WriteTimeout: pint(30),
			},
			wantDSN: "root:secret@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=30s&writeTimeout=30s",
		},
		{
			name: "missing host",
			config: Config{
				User:     "root",
				Password: "secret",
				Database: "mydb",
				Port:     3306,
			},
			wantError: ErrMysqlHostRequired,
		},
		{
			name: "missing user",
			config: Config{
				Host:     "localhost",
				Password: "secret",
				Database: "mydb",
				Port:     3306,
			},
			wantError: ErrMysqlUserRequired,
		},
		{
			name: "missing password",
			config: Config{
				Host:     "localhost",
				User:     "root",
				Database: "mydb",
				Port:     3306,
			},
			wantError: ErrMysqlPasswordRequired,
		},
		{
			name: "missing database",
			config: Config{
				Host:     "localhost",
				User:     "root",
				Password: "secret",
				Port:     3306,
			},
			wantError: ErrMysqlDatabaseRequired,
		},
		{
			name: "database port invalid (negative)",
			config: Config{
				Host:     "localhost",
				User:     "root",
				Password: "secret",
				Database: "mydb",
				Port:     -3306,
			},
			wantError: ErrMysqlInvalidPort,
		},
		{
			name: "database port invalid (too high)",
			config: Config{
				Host:     "localhost",
				User:     "root",
				Password: "secret",
				Database: "mydb",
				Port:     65536,
			},
			wantError: ErrMysqlInvalidPort,
		},
		{
			name: "set defayt port when value is not set",
			config: Config{
				Host:     "localhost",
				User:     "root",
				Password: "secret",
				Database: "mydb",
			},
			wantDSN: "root:secret@tcp(localhost:3306)/mydb",
		},
		{
			name: "invalid config: timeout with negative value",
			config: Config{
				Host:     "localhost",
				User:     "root",
				Password: "secret",
				Database: "mydb",
				Port:     3306,
				Timeout:  pint(-1),
			},
			wantError: ErrMysqlTimeoutInvalid,
		},
		{
			name: "invalid config: read_timeout with negative value",
			config: Config{
				Host:        "localhost",
				User:        "root",
				Password:    "secret",
				Database:    "mydb",
				Port:        3306,
				ReadTimeout: pint(-1),
			},
			wantError: ErrMysqlReadTimeoutInvalid,
		},
		{
			name: "invalid config: write_timeout with negative value",
			config: Config{
				Host:         "localhost",
				User:         "root",
				Password:     "secret",
				Database:     "mydb",
				Port:         3306,
				WriteTimeout: pint(-1),
			},
			wantError: ErrMysqlWriteTimeoutInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := tt.config.Build()
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

			if ds != tt.wantDSN {
				t.Errorf("dsn: got %s, want %s", ds, tt.wantDSN)
			}
		})
	}
}
