package oracle

import (
	"fmt"
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

// ConnectionString generates an Oracle Data Guard connection string.
// It validates the configuration and builds a TNS-style connection string with failover support.
func (c *DataGuardConfig) ConnectionString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	addressList := c.buildAddressList()
	connectData := c.buildConnectData()

	desc := fmt.Sprintf(
		"(DESCRIPTION=(ADDRESS_LIST=%s(FAILOVER=ON))(CONNECT_DATA=%s)%s)",
		addressList,
		connectData,
		c.buildTimeouts(),
	)

	return fmt.Sprintf("%s/%s@%s", c.User, c.Password, desc), nil
}

// buildAddressList constructs the ADDRESS_LIST section of the connection string.
// It includes the primary node and all standby nodes with their respective protocols, hosts, and ports.
func (c *DataGuardConfig) buildAddressList() string {
	var addresses []string

	primary := normalizeNode(c.Primary)
	addresses = append(addresses, fmt.Sprintf("(ADDRESS=(PROTOCOL=%s)(HOST=%s)(PORT=%d))", primary.Protocol, primary.Host, primary.Port))

	for _, standby := range c.Standbys {
		standby = normalizeNode(standby)
		addresses = append(addresses, fmt.Sprintf("(ADDRESS=(PROTOCOL=%s)(HOST=%s)(PORT=%d))", standby.Protocol, standby.Host, standby.Port))
	}

	return strings.Join(addresses, "")
}

// buildConnectData constructs the CONNECT_DATA section of the connection string.
// It includes the service name and optional failover configuration with retries and delay settings.
func (c *DataGuardConfig) buildConnectData() string {
	parts := []string{fmt.Sprintf("(SERVICE_NAME=%s)", c.ServiceName)}

	if c.FailoverMode != "" {
		failoverConfig := fmt.Sprintf("(FAILOVER_MODE=(TYPE=%s)", c.FailoverMode)

		if c.FailoverRetries > 0 {
			failoverConfig += fmt.Sprintf("(RETRIES=%d)", c.FailoverRetries)
		}

		if c.FailoverDelay > 0 {
			failoverConfig += fmt.Sprintf("(DELAY=%d)", c.FailoverDelay)
		}

		failoverConfig += ")"
		parts = append(parts, failoverConfig)
	}

	return strings.Join(parts, "")
}

// buildTimeouts constructs the timeout parameters section of the connection string.
// It includes CONNECT_TIMEOUT and TRANSPORT_CONNECT_TIMEOUT if they are configured.
func (c *DataGuardConfig) buildTimeouts() string {
	var parts []string

	if c.ConnectTimeout > 0 {
		parts = append(parts, fmt.Sprintf("(CONNECT_TIMEOUT=%d)", c.ConnectTimeout))
	}

	if c.TransportConnectTimeout > 0 {
		parts = append(parts, fmt.Sprintf("(TRANSPORT_CONNECT_TIMEOUT=%d)", c.TransportConnectTimeout))
	}

	return strings.Join(parts, "")
}
