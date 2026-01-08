package oracle

import (
	"fmt"
	"net/url"

	"github.com/pperesbr/gokit/pkg/dsn"
)

// StandaloneConfig represents the configuration for a standalone Oracle database connection.
// It supports both SERVICE_NAME and SID connection methods and includes timeout configurations.
type StandaloneConfig struct {
	// Host is the hostname or IP address of the Oracle database server.
	Host string
	// Port is the TCP port number of the Oracle database listener.
	// If zero, DefaultPort will be used.
	Port int
	// ServiceName is the Oracle service name for the connection.
	// Either ServiceName or SID must be specified, but not both.
	ServiceName string
	// SID is the Oracle System Identifier for the connection.
	// Either ServiceName or SID must be specified, but not both.
	SID string
	// Credentials contains the authentication information (User and Password).
	Credentials
	// Timeouts contains the connection timeout configurations.
	Timeouts
}

// ConnectionString builds and returns the Oracle connection string in go-ora URL format:
// oracle://user:password@host:port/service_name
// or oracle://user:password@host:port?SID=sid
// It validates the configuration before building the connection string.
func (s *StandaloneConfig) ConnectionString() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}

	port := s.Port
	if port == 0 {
		port = DefaultPort
	}

	u := &url.URL{
		Scheme: DriverName,
		User:   url.UserPassword(s.User, s.Password),
		Host:   fmt.Sprintf("%s:%d", s.Host, port),
	}

	// ServiceName goes in path, SID goes in query
	if s.ServiceName != "" {
		u.Path = s.ServiceName
	}

	q := u.Query()

	if s.SID != "" && s.ServiceName == "" {
		q.Set("SID", s.SID)
	}

	if s.ConnectTimeout > 0 {
		q.Set("TIMEOUT", fmt.Sprintf("%d", s.ConnectTimeout))
	}

	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// Validate checks if all required configuration fields are properly set.
// It validates the host, port range, service name or SID, user, and password.
// Returns a ValidationError if any required field is missing or invalid.
func (s *StandaloneConfig) Validate() error {
	if s.Host == "" {
		return dsn.NewValidationError(DriverName, "host", dsn.ErrMissingHost)
	}

	if s.Port != 0 && (s.Port < 1 || s.Port > 65535) {
		return dsn.NewValidationError(DriverName, "port", dsn.ErrInvalidPort)
	}

	if s.ServiceName == "" && s.SID == "" {
		return dsn.NewValidationError(DriverName, "service_name/sid", "service_name or sid is required")
	}

	if s.User == "" {
		return dsn.NewValidationError(DriverName, "user", dsn.ErrMissingUser)
	}

	if s.Password == "" {
		return dsn.NewValidationError(DriverName, "password", dsn.ErrMissingPassword)
	}

	return nil
}

// Driver returns the name of the Oracle database driver.
func (s *StandaloneConfig) Driver() string {
	return DriverName
}

var _ dsn.Builder = (*StandaloneConfig)(nil)
