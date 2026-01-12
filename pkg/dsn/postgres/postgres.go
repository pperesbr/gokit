// Package postgres provides PostgreSQL DSN (Data Source Name) configuration and builder functionality.
// It implements the dsn.DSN interface to construct valid PostgreSQL connection strings
// with support for various connection parameters including SSL modes, timeouts, and search paths.
package postgres

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/pperesbr/gokit/pkg/dsn"
)

var (
	_ dsn.DSN = (*Config)(nil)

	// validSSLModes contains the set of acceptable SSL mode values for PostgreSQL connections.
	validSSLModes = map[string]struct{}{
		"disable":     {},
		"allow":       {},
		"prefer":      {},
		"require":     {},
		"verify-ca":   {},
		"verify-full": {},
	}

	// ErrPostgresHostRequired is returned when the host field is empty.
	ErrPostgresHostRequired = errors.New("postgres: host is required")

	// ErrPostgresUserRequired is returned when the user field is empty.
	ErrPostgresUserRequired = errors.New("postgres: user is required")

	// ErrPostgresPasswordRequired is returned when the password field is empty.
	ErrPostgresPasswordRequired = errors.New("postgres: password is required")

	// ErrPostgresDatabaseRequired is returned when the database field is empty.
	ErrPostgresDatabaseRequired = errors.New("postgres: database is required")

	// ErrPostgresInvalidPort is returned when the port is not within the valid range of 1-65535.
	ErrPostgresInvalidPort = errors.New("postgres: port must between 1-65535")

	// ErrPostgresInvalidSSLMode is returned when an unsupported SSL mode value is provided.
	ErrPostgresInvalidSSLMode = errors.New("postgres: invalid sslmode value, valid values are: disable, allow, prefer, require, verify-ca, verify-full")

	// ErrPostgresInvalidConnectTimeout is returned when the connect_timeout value is negative.
	ErrPostgresInvalidConnectTimeout = errors.New("postgres: connect_timeout must be greater than or equal to 0")
)

// Config holds the configuration parameters required to build a PostgreSQL DSN.
// It supports all standard PostgreSQL connection parameters including SSL configuration,
// application identification, connection timeouts, and schema/timezone settings.
type Config struct {
	// Host specifies the PostgreSQL server hostname or IP address.
	Host string `yaml:"host"`

	// User specifies the PostgreSQL username for authentication.
	User string `yaml:"user"`

	// Password specifies the password for the PostgreSQL user.
	Password string `yaml:"password"`

	// Database specifies the name of the PostgreSQL database to connect to.
	Database string `yaml:"database"`

	// Port specifies the PostgreSQL server port. Defaults to 5432 if not set or zero.
	Port int `yaml:"port"`

	// SSLMode specifies the SSL/TLS connection mode. Valid values are:
	// disable, allow, prefer, require, verify-ca, verify-full.
	SSLMode string `yaml:"ssl_mode"`

	// ApplicationName specifies the name of the application connecting to the database.
	// This value appears in PostgreSQL logs and statistics views.
	ApplicationName string `yaml:"application_name"`

	// ConnectTimeout specifies the maximum time in seconds to wait for a connection.
	// If nil or negative, no timeout is applied. Must be >= 0 if set.
	ConnectTimeout *int `yaml:"connection_timeout"`

	// SearchPath specifies the schema search path for the connection.
	SearchPath string `yaml:"search_path"`

	// Timezone specifies the timezone to use for the connection.
	Timezone string `yaml:"timezone"`
}

// Build constructs a PostgreSQL DSN connection string from the Config parameters.
// It validates all required fields and optional parameters before building the DSN.
// The resulting DSN follows the format: postgres://user:password@host:port/database?params
//
// Returns an error if any required field is missing or if any parameter is invalid.
func (c *Config) Build() (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}

	var params []string
	if c.SSLMode != "" {
		params = append(params, fmt.Sprintf("sslmode=%s", c.SSLMode))
	}

	if c.ApplicationName != "" {
		params = append(params, fmt.Sprintf("application_name=%s", url.QueryEscape(c.ApplicationName)))
	}

	if c.ConnectTimeout != nil && *c.ConnectTimeout >= 0 {
		params = append(params, fmt.Sprintf("connect_timeout=%d", *c.ConnectTimeout))
	}

	if c.SearchPath != "" {
		params = append(params, fmt.Sprintf("search_path=%s", url.QueryEscape(c.SearchPath)))
	}

	if c.Timezone != "" {
		params = append(params, fmt.Sprintf("timezone=%s", url.QueryEscape(c.Timezone)))
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		c.Host,
		c.Port,
		c.Database,
	)

	if len(params) > 0 {
		dsn = dsn + "?" + strings.Join(params, "&")
	}

	return dsn, nil

}

// validate checks that all required fields are present and all parameters have valid values.
// It sets the default port to 5432 if not specified. Returns an error if validation fails.
func (c *Config) validate() error {
	if c.Host == "" {
		return ErrPostgresHostRequired
	}

	if c.User == "" {
		return ErrPostgresUserRequired
	}

	if c.Password == "" {
		return ErrPostgresPasswordRequired
	}

	if c.Database == "" {
		return ErrPostgresDatabaseRequired
	}

	if c.Port == 0 {
		c.Port = 5432
	}

	if c.Port < 0 || c.Port > 65535 {
		return ErrPostgresInvalidPort
	}

	if c.SSLMode != "" && !isValidSSLMode(c.SSLMode) {
		return ErrPostgresInvalidSSLMode
	}

	if c.ConnectTimeout != nil && *c.ConnectTimeout < 0 {
		return ErrPostgresInvalidConnectTimeout
	}

	return nil
}

// isValidSSLMode checks if the provided SSL mode string is one of the valid PostgreSQL SSL modes.
func isValidSSLMode(mode string) bool {
	_, ok := validSSLModes[mode]
	return ok
}
