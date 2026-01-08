// Package mysql provides configuration and connection string building
// for MySQL database connections.
package mysql

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pperesbr/gokit/pkg/dsn"
	"gopkg.in/yaml.v3"
)

const (
	// DriverName is the name of the MySQL driver.
	DriverName = "mysql"
	// DefaultPort is the default MySQL port.
	DefaultPort = 3306
	// DefaultCharset is the default charset for connections.
	DefaultCharset = "utf8mb4"
	// DefaultProtocol is the default network protocol.
	DefaultProtocol = "tcp"
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
	// Timeout is the connection timeout in seconds.
	Timeout int `yaml:"timeout"`
	// ReadTimeout is the I/O read timeout in seconds.
	ReadTimeout int `yaml:"read_timeout"`
	// WriteTimeout is the I/O write timeout in seconds.
	WriteTimeout int `yaml:"write_timeout"`
}

// Config represents the configuration for a MySQL database connection.
type Config struct {
	// Host is the hostname or IP address of the MySQL server.
	Host string `yaml:"host"`
	// Port is the TCP port number of the MySQL server.
	// If zero, DefaultPort will be used.
	Port int `yaml:"port"`
	// Database is the name of the database to connect to.
	Database string `yaml:"database"`
	// Protocol is the network protocol (tcp, unix).
	// If empty, DefaultProtocol will be used.
	Protocol string `yaml:"protocol"`
	// Charset is the character set for the connection.
	// If empty, DefaultCharset will be used.
	Charset string `yaml:"charset"`
	// ParseTime enables parsing of DATE and DATETIME values to time.Time.
	ParseTime bool `yaml:"parse_time"`
	// Loc is the location for time.Time values (e.g., "Local", "UTC").
	Loc string `yaml:"loc"`
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
		return nil, fmt.Errorf("failed to parse mysql config: %w", err)
	}

	return &cfg, nil
}

// ConnectionString builds and returns the MySQL connection string in DSN format:
// user:password@tcp(host:port)/database?charset=utf8mb4&parseTime=true
// It validates the configuration before building the connection string.
func (c *Config) ConnectionString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	port := c.Port
	if port == 0 {
		port = DefaultPort
	}

	protocol := c.Protocol
	if protocol == "" {
		protocol = DefaultProtocol
	}

	charset := c.Charset
	if charset == "" {
		charset = DefaultCharset
	}

	// Build DSN: user:password@protocol(host:port)/database?params
	var sb strings.Builder

	// user:password@
	sb.WriteString(url.QueryEscape(c.User))
	sb.WriteString(":")
	sb.WriteString(url.QueryEscape(c.Password))
	sb.WriteString("@")

	// protocol(host:port)
	sb.WriteString(protocol)
	sb.WriteString("(")
	sb.WriteString(c.Host)
	sb.WriteString(":")
	sb.WriteString(fmt.Sprintf("%d", port))
	sb.WriteString(")")

	// /database
	sb.WriteString("/")
	sb.WriteString(c.Database)

	// ?params
	params := c.buildParams(charset)
	if len(params) > 0 {
		sb.WriteString("?")
		sb.WriteString(params)
	}

	return sb.String(), nil
}

// buildParams builds the query parameters for the connection string.
func (c *Config) buildParams(charset string) string {
	params := make([]string, 0)

	params = append(params, "charset="+charset)

	if c.ParseTime {
		params = append(params, "parseTime=true")
	}

	if c.Loc != "" {
		params = append(params, "loc="+url.QueryEscape(c.Loc))
	}

	if c.Timeout > 0 {
		params = append(params, fmt.Sprintf("timeout=%ds", c.Timeout))
	}

	if c.ReadTimeout > 0 {
		params = append(params, fmt.Sprintf("readTimeout=%ds", c.ReadTimeout))
	}

	if c.WriteTimeout > 0 {
		params = append(params, fmt.Sprintf("writeTimeout=%ds", c.WriteTimeout))
	}

	return strings.Join(params, "&")
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

	if c.Protocol != "" && !isValidProtocol(c.Protocol) {
		return dsn.NewValidationError(DriverName, "protocol", "must be one of: tcp, unix")
	}

	return nil
}

// Driver returns the name of the MySQL database driver.
func (c *Config) Driver() string {
	return DriverName
}

// isValidProtocol checks if the provided protocol is valid.
func isValidProtocol(protocol string) bool {
	switch protocol {
	case "tcp", "unix":
		return true
	default:
		return false
	}
}

var _ dsn.Builder = (*Config)(nil)
