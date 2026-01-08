package oracle

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pperesbr/gokit/pkg/dsn"
)

var _ dsn.Builder = (*DataGuardConfig)(nil)

const (
	// FailoverModeSession defines session-level failover mode for Oracle Data Guard.
	FailoverModeSession = "SESSION"
	// FailoverModeSelect defines select-level failover mode for Oracle Data Guard.
	FailoverModeSelect = "SELECT"
)

// DataGuardConfig represents the configuration for Oracle Data Guard with primary and standby nodes.
// It includes connection details, failover settings, and timeout configurations.
type DataGuardConfig struct {
	// Primary is the primary database node.
	Primary Node
	// Standbys is the list of standby database nodes for failover.
	Standbys []Node
	// ServiceName is the Oracle service name to connect to.
	ServiceName string
	// Credentials contains the username and password for authentication.
	Credentials
	// Timeouts contains connection timeout configurations.
	Timeouts
	// FailoverMode specifies the failover mode (SESSION or SELECT).
	FailoverMode string
	// FailoverRetries specifies the number of failover retry attempts.
	FailoverRetries int
	// FailoverDelay specifies the delay in seconds between failover retries.
	FailoverDelay int
}

// Driver returns the Oracle driver name.
func (c *DataGuardConfig) Driver() string {
	return DriverName
}

// Validate checks if the Data Guard configuration is valid.
// It verifies that all required fields are present and valid, including primary and standby nodes,
// service name, credentials, and failover mode settings.
func (c *DataGuardConfig) Validate() error {
	if c.Primary.Host == "" {
		return dsn.NewValidationError(DriverName, "primary.host", dsn.ErrMissingHost)
	}

	if c.Primary.Port != 0 && (c.Primary.Port < 1 || c.Primary.Port > 65535) {
		return dsn.NewValidationError(DriverName, "primary.port", dsn.ErrInvalidPort)
	}

	if len(c.Standbys) == 0 {
		return dsn.NewValidationError(DriverName, "standbys", "at least one standby is required")
	}

	for i, standby := range c.Standbys {
		if standby.Host == "" {
			return dsn.NewValidationError(DriverName, fmt.Sprintf("standbys[%d].host", i), dsn.ErrMissingHost)
		}

		if standby.Port != 0 && (standby.Port < 1 || standby.Port > 65535) {
			return dsn.NewValidationError(DriverName, fmt.Sprintf("standbys[%d].port", i), dsn.ErrInvalidPort)
		}
	}

	if c.ServiceName == "" {
		return dsn.NewValidationError(DriverName, "service_name", "is required")
	}

	if c.User == "" {
		return dsn.NewValidationError(DriverName, "user", dsn.ErrMissingUser)
	}

	if c.Password == "" {
		return dsn.NewValidationError(DriverName, "password", dsn.ErrMissingPassword)
	}

	if c.FailoverMode != "" && c.FailoverMode != FailoverModeSession && c.FailoverMode != FailoverModeSelect {
		return dsn.NewValidationError(DriverName, "failover_mode", "must be SESSION or SELECT")
	}

	return nil
}

// ConnectionString generates an Oracle Data Guard connection string in go-ora URL format.
// Format: oracle://user:password@primary:port,standby1:port/service_name?FAILOVER=true&FAILOVER_MODE=SESSION
// It validates the configuration and builds a URL with primary and standby hosts.
func (c *DataGuardConfig) ConnectionString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	// Build hosts list (primary + standbys)
	hosts := c.buildHostsList()

	u := &url.URL{
		Scheme: DriverName,
		User:   url.UserPassword(c.User, c.Password),
		Host:   hosts,
		Path:   c.ServiceName,
	}

	q := u.Query()

	// Data Guard always has failover enabled
	q.Set("FAILOVER", "true")

	if c.FailoverMode != "" {
		q.Set("FAILOVER_MODE", c.FailoverMode)
	}

	if c.FailoverRetries > 0 {
		q.Set("FAILOVER_RETRIES", fmt.Sprintf("%d", c.FailoverRetries))
	}

	if c.FailoverDelay > 0 {
		q.Set("FAILOVER_DELAY", fmt.Sprintf("%d", c.FailoverDelay))
	}

	if c.ConnectTimeout > 0 {
		q.Set("TIMEOUT", fmt.Sprintf("%d", c.ConnectTimeout))
	}

	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// buildHostsList constructs the comma-separated list of hosts for Data Guard connection.
// Primary node comes first, followed by standby nodes.
func (c *DataGuardConfig) buildHostsList() string {
	var hosts []string

	primary := normalizeNode(c.Primary)
	hosts = append(hosts, fmt.Sprintf("%s:%d", primary.Host, primary.Port))

	for _, standby := range c.Standbys {
		standby = normalizeNode(standby)
		hosts = append(hosts, fmt.Sprintf("%s:%d", standby.Host, standby.Port))
	}

	return strings.Join(hosts, ",")
}
