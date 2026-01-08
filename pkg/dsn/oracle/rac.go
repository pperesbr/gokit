package oracle

import (
	"fmt"
	"net/url"
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

// ConnectionString generates the Oracle RAC connection string in go-ora URL format.
// Format: oracle://user:password@host1:port1,host2:port2/service_name?FAILOVER=true&LOAD_BALANCE=true
// It validates the configuration and builds a URL with multiple hosts.
func (c *RACConfig) ConnectionString() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	// Build hosts list
	hosts := c.buildHostsList()

	u := &url.URL{
		Scheme: DriverName,
		User:   url.UserPassword(c.User, c.Password),
		Host:   hosts,
		Path:   c.ServiceName,
	}

	q := u.Query()

	if c.LoadBalance {
		q.Set("LOAD_BALANCE", "true")
	}

	if c.Failover {
		q.Set("FAILOVER", "true")
	}

	if c.RetryCount > 0 {
		q.Set("RETRY_COUNT", fmt.Sprintf("%d", c.RetryCount))
	}

	if c.RetryDelay > 0 {
		q.Set("RETRY_DELAY", fmt.Sprintf("%d", c.RetryDelay))
	}

	if c.ConnectTimeout > 0 {
		q.Set("TIMEOUT", fmt.Sprintf("%d", c.ConnectTimeout))
	}

	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// buildHostsList constructs the comma-separated list of hosts for RAC connection.
func (c *RACConfig) buildHostsList() string {
	var hosts []string

	for _, node := range c.Nodes {
		node = normalizeNode(node)
		hosts = append(hosts, fmt.Sprintf("%s:%d", node.Host, node.Port))
	}

	return strings.Join(hosts, ",")
}
