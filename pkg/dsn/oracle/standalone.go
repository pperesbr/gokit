package oracle

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pperesbr/gokit/pkg/dsn"
)

var _ dsn.DSN = (*StandaloneConfig)(nil)

// StandaloneConfig represents the configuration for a standalone Oracle database connection.
// It implements the dsn.DSN interface and provides methods to build and validate
// Oracle connection strings.
type StandaloneConfig struct {
	// Host specifies the hostname or IP address of the Oracle database server.
	Host string `yaml:"host"`

	// User specifies the username for authenticating to the Oracle database.
	User string `yaml:"user"`

	// Password specifies the password for authenticating to the Oracle database.
	Password string `yaml:"password"`

	// Port specifies the TCP port number on which the Oracle database is listening.
	// Defaults to 1521 if not specified or set to 0.
	Port int `yaml:"port"`

	// ServiceName specifies the Oracle service name to connect to.
	ServiceName string `yaml:"service_name"`

	// ConnectionTimeout specifies the connection timeout in seconds.
	// Optional field; if nil, no connection timeout is set.
	ConnectionTimeout *int `yaml:"connection_timeout"`

	// Timeout specifies the general operation timeout in seconds.
	// Optional field; if nil, no timeout is set.
	Timeout *int `yaml:"timeout"`
}

// Build constructs and returns an Oracle DSN string from the StandaloneConfig.
// It validates the configuration first, then builds a connection string in the format:
// oracle://user:password@host:port/service_name?params
// Returns an error if validation fails.
func (s *StandaloneConfig) Build() (string, error) {
	if err := s.validate(); err != nil {
		return "", err
	}

	var params []string

	if s.ConnectionTimeout != nil {
		params = append(params, fmt.Sprintf("CONNECTION TIMEOUT=%d", *s.ConnectionTimeout))
	}

	if s.Timeout != nil {
		params = append(params, fmt.Sprintf("TIMEOUT=%d", *s.Timeout))
	}

	dsn := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		url.QueryEscape(s.User),
		url.QueryEscape(s.Password),
		s.Host,
		s.Port,
		url.QueryEscape(s.ServiceName),
	)

	if len(params) > 0 {
		dsn = dsn + "?" + strings.Join(params, "&")
	}

	return dsn, nil

}

// validate checks that all required fields are set and contain valid values.
// It sets default values where appropriate (e.g., Port defaults to 1521).
// Returns an error if any validation check fails.
func (s *StandaloneConfig) validate() error {
	if s.Host == "" {
		return ErrOracleHostRequired
	}

	if s.User == "" {
		return ErrOracleUserRequired
	}

	if s.Password == "" {
		return ErrOraclePasswordRequired
	}

	if s.ServiceName == "" {
		return ErrOracleServiceNameRequired
	}

	if s.Port == 0 {
		s.Port = 1521
	}

	if s.Port < 0 || s.Port > 65535 {
		return ErrOraclePortInvalid
	}

	if s.ConnectionTimeout != nil && *s.ConnectionTimeout < 0 {
		return ErrOracleConnectTimeoutInvalid
	}

	if s.Timeout != nil && *s.Timeout < 0 {
		return ErrOracleTimeoutInvalid
	}

	return nil
}
