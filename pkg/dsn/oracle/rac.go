package oracle

import (
	"fmt"
	"strings"

	"github.com/pperesbr/gokit/pkg/dsn"
)

var _ dsn.Builder = (*RACConfig)(nil)

// RACConfig represents the configuration for Oracle Real Application Clusters (RAC).
// It implements the dsn.Builder interface to generate connection strings for RAC environments.
type RACConfig struct {
	// Nodes is the list of Oracle RAC nodes to connect to.
	Nodes []Node
	// ServiceName is the Oracle service name to connect to.
	ServiceName string
	Credentials
	Timeouts
	// LoadBalance indicates whether to enable load balancing across RAC nodes.
	LoadBalance bool
	// Failover indicates whether to enable automatic failover between RAC nodes.
	Failover bool
	// RetryCount is the number of connection retry attempts.
	RetryCount int
	// RetryDelay is the delay in seconds between connection retry attempts.
	RetryDelay int
}

// Driver returns the driver name for Oracle RAC connections.
func (c *RACConfig) Driver() string {
	return DriverName
}

// Validate checks if the RAC configuration is valid.
// It ensures that all required fields are set and have valid values.
func (c *RACConfig) Validate() error {
	if len(c.Nodes) == 0 {
		return dsn.NewValidationError(DriverName, "nodes", "at least one node is required")
	}

	for i, node := range c.Nodes {
		if node.Host == "" {
			return dsn.NewValidationError(DriverName, fmt.Sprintf("nodes[%d].host", i), dsn.ErrMissingHost)
		}

		if node.Port != 0 && (node.Port < 1 || node.Port > 65535) {
			return dsn.NewValidationError(DriverName, fmt.Sprintf("nodes[%d].port", i), dsn.ErrInvalidPort)
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

	return nil
}

// ConnectionString generates the Oracle RAC connection string.
// It validates the configuration and builds a TNS descriptor with multiple addresses,
// load balancing, failover, and timeout settings.
func (c *RACConfig) ConnectionString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	addressList := c.buildAddressList()
	connectData := fmt.Sprintf("(SERVICE_NAME=%s)", c.ServiceName)

	desc := fmt.Sprintf(
		"(DESCRIPTION=(ADDRESS_LIST=%s%s)(CONNECT_DATA=%s)%s)",
		addressList,
		c.buildLoadBalanceFailover(),
		connectData,
		c.buildTimeouts(),
	)

	return fmt.Sprintf("%s/%s@%s", c.User, c.Password, desc), nil
}

// buildAddressList constructs the ADDRESS_LIST section of the TNS descriptor
// by iterating through all configured RAC nodes.
func (c *RACConfig) buildAddressList() string {
	var addresses []string

	for _, node := range c.Nodes {
		node = normalizeNode(node)
		addr := fmt.Sprintf("(ADDRESS=(PROTOCOL=%s)(HOST=%s)(PORT=%d))", node.Protocol, node.Host, node.Port)
		addresses = append(addresses, addr)
	}

	return strings.Join(addresses, "")
}

// buildLoadBalanceFailover constructs the load balancing and failover parameters
// for the TNS descriptor, including retry count and delay settings.
func (c *RACConfig) buildLoadBalanceFailover() string {
	var parts []string

	if c.LoadBalance {
		parts = append(parts, "(LOAD_BALANCE=ON)")
	}

	if c.Failover {
		parts = append(parts, "(FAILOVER=ON)")
	}

	if c.RetryCount > 0 {
		parts = append(parts, fmt.Sprintf("(RETRY_COUNT=%d)", c.RetryCount))
	}

	if c.RetryDelay > 0 {
		parts = append(parts, fmt.Sprintf("(RETRY_DELAY=%d)", c.RetryDelay))
	}

	return strings.Join(parts, "")
}

// buildTimeouts constructs the timeout parameters for the TNS descriptor,
// including connect timeout and transport connect timeout settings.
func (c *RACConfig) buildTimeouts() string {
	var parts []string

	if c.ConnectTimeout > 0 {
		parts = append(parts, fmt.Sprintf("(CONNECT_TIMEOUT=%d)", c.ConnectTimeout))
	}

	if c.TransportConnectTimeout > 0 {
		parts = append(parts, fmt.Sprintf("(TRANSPORT_CONNECT_TIMEOUT=%d)", c.TransportConnectTimeout))
	}

	return strings.Join(parts, "")
}
