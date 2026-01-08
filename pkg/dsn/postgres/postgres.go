// Package postgres provides configuration and connection string building
// for PostgreSQL database connections.
package postgres

import (
	"fmt"
	"net/url"

	"github.com/pperesbr/gokit/pkg/dsn"
	"gopkg.in/yaml.v3"
)

var _ dsn.Builder = (*Config)(nil)

const (
	// DriverName is the name of the PostgreSQL driver.
	DriverName = "postgres"
	// DefaultPort is the default PostgreSQL port.
	DefaultPort = 5432
	// DefaultSSLMode is the default SSL mode for connections.
	DefaultSSLMode = "disable"
)

// Credentials contains the authentication information for the database connection.
type Credentials struct {
	// User is the username for authentication.
	User string `yaml:"user"`
	// Password is the password for authentication.
	Password string `yaml:"password"`
}

// Timeouts contains the timeout configurations for the database connection.
type Timeouts struct {
	// ConnectTimeout is the maximum wait time for connection in seconds.
	// If zero, no timeout is applied.
	ConnectTimeout int `yaml:"connect_timeout"`
}

// Config represents the configuration for a PostgreSQL database connection.
type Config struct {
	// Host is the hostname or IP address of the PostgreSQL server.
	Host string `yaml:"host"`
	// Port is the TCP port number of the PostgreSQL server.
	// If zero, DefaultPort will be used.
	Port int `yaml:"port"`
	// Database is the name of the database to connect to.
	Database string `yaml:"database"`
	// SSLMode defines the SSL connection mode.
	// Valid values: disable, require, verify-ca, verify-full.
	// If empty, DefaultSSLMode will be used.
	SSLMode string `yaml:"sslmode"`
	// Credentials contains the authentication information (User and Password).
	Credentials `yaml:",inline"`
	// Timeouts contains the connection timeout configurations.
	Timeouts `yaml:",inline"`
}

// NewBuilder creates a new DSN builder from YAML configuration data.
// It parses the provided YAML data and returns a Config instance ready to build connection strings.
// Returns an error if the YAML cannot be parsed.
func NewBuilder(data []byte) (dsn.Builder, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	return &cfg, nil
}

// ConnectionString builds and returns the PostgreSQL connection string in URL format:
// postgres://user:password@host:port/database?sslmode=disable&connect_timeout=10
// It validates the configuration before building the connection string.
func (c *Config) ConnectionString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	port := c.Port
	if port == 0 {
		port = DefaultPort
	}

	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = DefaultSSLMode
	}

	// Build URL with proper escaping
	u := &url.URL{
		Scheme: DriverName,
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, port),
		Path:   c.Database,
	}

	// Add query parameters
	q := u.Query()
	q.Set("sslmode", sslMode)

	if c.ConnectTimeout > 0 {
		q.Set("connect_timeout", fmt.Sprintf("%d", c.ConnectTimeout))
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

// Validate checks if all required configuration fields are properly set.
// It validates the host, port range, database, user, and password.
// Returns a ValidationError if any required field is missing or invalid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return dsn.NewValidationError(DriverName, "host", dsn.ErrMissingHost)
	}

	if c.Port != 0 && (c.Port < 1 || c.Port > 65535) {
		return dsn.NewValidationError(DriverName, "port", dsn.ErrInvalidPort)
	}

	if c.Database == "" {
		return dsn.NewValidationError(DriverName, "database", dsn.ErrMissingDatabase)
	}

	if c.User == "" {
		return dsn.NewValidationError(DriverName, "user", dsn.ErrMissingUser)
	}

	if c.Password == "" {
		return dsn.NewValidationError(DriverName, "password", dsn.ErrMissingPassword)
	}

	if c.SSLMode != "" && !isValidSSLMode(c.SSLMode) {
		return dsn.NewValidationError(DriverName, "sslmode", "must be one of: disable, require, verify-ca, verify-full")
	}

	return nil
}

// Driver returns the name of the PostgreSQL database driver.
func (c *Config) Driver() string {
	return DriverName
}

// isValidSSLMode checks if the provided SSL mode is valid.
func isValidSSLMode(mode string) bool {
	switch mode {
	case "disable", "require", "verify-ca", "verify-full":
		return true
	default:
		return false
	}
}
