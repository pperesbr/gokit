// Package mysql provides MySQL DSN (Data Source Name) configuration and building functionality.
// It implements the dsn.DSN interface to construct valid MySQL connection strings.
package mysql

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/pperesbr/gokit/pkg/dsn"
)

var (
	_ dsn.DSN = (*Config)(nil)

	ErrMysqlHostRequired        = errors.New("mysql: host is required")
	ErrMysqlUserRequired        = errors.New("mysql: user is required")
	ErrMysqlPasswordRequired    = errors.New("mysql: password is required")
	ErrMysqlDatabaseRequired    = errors.New("mysql: database is required")
	ErrMysqlInvalidPort         = errors.New("mysql: port must between 1-65535")
	ErrMysqlTimeoutInvalid      = errors.New("mysql: timeout must be greater than or equal to 0")
	ErrMysqlReadTimeoutInvalid  = errors.New("mysql: readTimeout must be greater than or equal to 0")
	ErrMysqlWriteTimeoutInvalid = errors.New("mysql: writeTimeout must be greater than or equal to 0")
)

// Config represents the MySQL database connection configuration.
// It contains all necessary parameters to build a valid MySQL DSN string.
type Config struct {
	// Host is the MySQL server hostname or IP address (required).
	Host string `yaml:"host"`
	// User is the MySQL username for authentication (required).
	User string `yaml:"user"`
	// Password is the MySQL password for authentication (required).
	Password string `yaml:"password"`
	// Database is the name of the database to connect to (required).
	Database string `yaml:"database"`
	// Port is the MySQL server port (defaults to 3306 if not specified).
	Port int `yaml:"port"`
	// Charset specifies the character set for the connection (optional).
	Charset string `yaml:"charset"`
	// ParseTime determines whether to parse time values to time.Time (optional).
	ParseTime *bool `yaml:"parseTime"`
	// Loc specifies the location for time.Time values (optional).
	Loc string `yaml:"loc"`
	// Timeout specifies the connection timeout in seconds (optional, must be >= 0).
	Timeout *int `yaml:"timeout"`
	// ReadTimeout specifies the I/O read timeout in seconds (optional, must be >= 0).
	ReadTimeout *int `yaml:"readTimeout"`
	// WriteTimeout specifies the I/O write timeout in seconds (optional, must be >= 0).
	WriteTimeout *int `yaml:"writeTimeout"`
}

// Build constructs and returns a MySQL DSN string from the configuration.
// It validates the configuration first and returns an error if validation fails.
// The returned DSN string follows the format: user:password@tcp(host:port)/database?params
func (c *Config) Build() (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}

	var params []string
	if c.Charset != "" {
		params = append(params, fmt.Sprintf("charset=%s", url.QueryEscape(c.Charset)))
	}

	if c.ParseTime != nil {
		valueStr := "True"

		if !*c.ParseTime {
			valueStr = "False"
		}

		params = append(params, fmt.Sprintf("parseTime=%s", valueStr))
	}

	if c.Loc != "" {
		params = append(params, fmt.Sprintf("loc=%s", url.QueryEscape(c.Loc)))
	}

	if c.Timeout != nil {
		params = append(params, fmt.Sprintf("timeout=%ds", *c.Timeout))
	}

	if c.ReadTimeout != nil {
		params = append(params, fmt.Sprintf("readTimeout=%ds", *c.ReadTimeout))
	}

	if c.WriteTimeout != nil {
		params = append(params, fmt.Sprintf("writeTimeout=%ds", *c.WriteTimeout))
	}

	dsn := fmt.Sprintf(""+
		"%s:%s@tcp(%s:%d)/%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		c.Host,
		c.Port,
		c.Database)

	if len(params) > 0 {
		dsn = dsn + "?" + strings.Join(params, "&")
	}

	return dsn, nil

}

// validate checks if all required configuration fields are properly set.
// It ensures Host, User, Password, and Database are not empty.
// It also validates Port is within valid range (1-65535), defaulting to 3306 if zero.
// Timeout values (Timeout, ReadTimeout, WriteTimeout) must be non-negative if provided.
func (c *Config) validate() error {
	if c.Host == "" {
		return ErrMysqlHostRequired
	}

	if c.User == "" {
		return ErrMysqlUserRequired
	}

	if c.Password == "" {
		return ErrMysqlPasswordRequired
	}

	if c.Database == "" {
		return ErrMysqlDatabaseRequired
	}

	if c.Port == 0 {
		c.Port = 3306
	}

	if c.Port < 1 || c.Port > 65535 {
		return ErrMysqlInvalidPort
	}

	if c.Timeout != nil && *c.Timeout < 0 {
		return ErrMysqlTimeoutInvalid
	}

	if c.ReadTimeout != nil && *c.ReadTimeout < 0 {
		return ErrMysqlReadTimeoutInvalid
	}

	if c.WriteTimeout != nil && *c.WriteTimeout < 0 {
		return ErrMysqlWriteTimeoutInvalid
	}

	return nil
}
