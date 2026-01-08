package oracle

import (
	"fmt"
	"strings"

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

// ConnectionString builds and returns the Oracle connection string in the format:
// user/password@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host)(PORT=port))(CONNECT_DATA=...)(TIMEOUTS...))
// It validates the configuration before building the connection string.
func (s *StandaloneConfig) ConnectionString() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}

	port := s.Port
	if port == 0 {
		port = DefaultPort
	}

	connectData := s.buildConnectData()

	desc := fmt.Sprintf(
		"(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=%s)(PORT=%d))(CONNECT_DATA=%s)%s)",
		s.Host,
		port,
		connectData,
		s.buildTimeouts(),
	)

	return fmt.Sprintf("%s/%s@%s", s.User, s.Password, desc), nil
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

// buildConnectData builds the CONNECT_DATA portion of the connection descriptor.
// It returns either (SERVICE_NAME=...) or (SID=...) based on which field is set.
// ServiceName takes precedence over SID if both are specified.
func (c *StandaloneConfig) buildConnectData() string {
	if c.ServiceName != "" {
		return fmt.Sprintf("(SERVICE_NAME=%s)", c.ServiceName)
	}
	return fmt.Sprintf("(SID=%s)", c.SID)
}

// buildTimeouts builds the timeout parameters portion of the connection descriptor.
// It includes CONNECT_TIMEOUT and TRANSPORT_CONNECT_TIMEOUT if they are greater than zero.
// Returns an empty string if no timeouts are configured.
func (c *StandaloneConfig) buildTimeouts() string {
	var parts []string

	if c.ConnectTimeout > 0 {
		parts = append(parts, fmt.Sprintf("(CONNECT_TIMEOUT=%d)", c.ConnectTimeout))
	}

	if c.TransportConnectTimeout > 0 {
		parts = append(parts, fmt.Sprintf("(TRANSPORT_CONNECT_TIMEOUT=%d)", c.TransportConnectTimeout))
	}

	return strings.Join(parts, "")
}

// Driver returns the name of the Oracle database driver.
func (s *StandaloneConfig) Driver() string {
	return DriverName
}

var _ dsn.Builder = (*StandaloneConfig)(nil)
