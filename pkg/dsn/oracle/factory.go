// Package oracle provides DSN (Data Source Name) builders for Oracle database connections.
// It supports three connection modes: standalone, RAC (Real Application Clusters), and DataGuard.
package oracle

import (
	"fmt"

	"github.com/pperesbr/gokit/pkg/dsn"
	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure for Oracle database connections.
// It supports three modes: "standalone" for single instance, "rac" for Real Application Clusters,
// and "dataguard" for Data Guard configurations. The configuration is typically loaded from YAML.
type Config struct {
	// Mode determines the connection mode: "standalone", "rac", or "dataguard"
	Mode string `yaml:"mode"`

	// Standalone
	// Host is the hostname or IP address for standalone database connections
	Host string `yaml:"host"`
	// Port is the port number for standalone database connections
	Port int `yaml:"port"`
	// ServiceName is the Oracle service name used for standalone and RAC connections
	ServiceName string `yaml:"service_name"`
	// SID is the Oracle System Identifier for standalone connections (alternative to ServiceName)
	SID string `yaml:"sid"`

	// RAC
	// Nodes is the list of Oracle RAC cluster nodes
	Nodes []NodeConfig `yaml:"nodes"`
	// LoadBalance enables load balancing across RAC nodes when true
	LoadBalance bool `yaml:"load_balance"`
	// Failover enables automatic failover to other RAC nodes when true
	Failover bool `yaml:"failover"`
	// RetryCount specifies the number of connection retry attempts for RAC
	RetryCount int `yaml:"retry_count"`
	// RetryDelay specifies the delay in seconds between retry attempts for RAC
	RetryDelay int `yaml:"retry_delay"`

	// DataGuard
	// Primary is the primary database node configuration for DataGuard
	Primary NodeConfig `yaml:"primary"`
	// Standbys is the list of standby database nodes for DataGuard
	Standbys []NodeConfig `yaml:"standbys"`
	// FailoverMode specifies the failover mode for DataGuard (e.g., "select", "session")
	FailoverMode string `yaml:"failover_mode"`
	// FailoverRetries specifies the number of failover retry attempts for DataGuard
	FailoverRetries int `yaml:"failover_retries"`
	// FailoverDelay specifies the delay in seconds between failover attempts for DataGuard
	FailoverDelay int `yaml:"failover_delay"`

	// Common
	// User is the database username for authentication
	User string `yaml:"user"`
	// Password is the database password for authentication
	Password string `yaml:"password"`
	// ConnectTimeout is the connection timeout in seconds
	ConnectTimeout int `yaml:"connect_timeout"`
	// TransportConnectTimeout is the transport layer connection timeout in seconds
	TransportConnectTimeout int `yaml:"transport_connect_timeout"`
}

// NodeConfig represents a single Oracle database node configuration.
// It is used for RAC and DataGuard setups to define individual nodes in the cluster.
type NodeConfig struct {
	// Host is the hostname or IP address of the database node
	Host string `yaml:"host"`
	// Port is the port number of the database node
	Port int `yaml:"port"`
	// Protocol is the connection protocol for the node (e.g., "TCP", "TCPS")
	Protocol string `yaml:"protocol"`
}

// NewBuilder creates a new DSN builder from YAML configuration data.
// It parses the provided YAML data and returns the appropriate builder based on the mode field.
// Supported modes are "standalone" (or empty), "rac", and "dataguard".
// Returns an error if the YAML cannot be parsed or if an unsupported mode is specified.
func NewBuilder(data []byte) (dsn.Builder, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse oracle config: %w", err)
	}

	switch cfg.Mode {
	case "standalone", "":
		return newStandaloneFromConfig(cfg), nil
	case "rac":
		return newRACFromConfig(cfg), nil
	case "dataguard":
		return newDataGuardFromConfig(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported oracle mode: %s", cfg.Mode)
	}
}

// newStandaloneFromConfig creates a StandaloneConfig from the generic Config structure.
// It extracts standalone-specific fields and common fields to build a standalone Oracle configuration.
func newStandaloneFromConfig(cfg Config) *StandaloneConfig {
	return &StandaloneConfig{
		Host:        cfg.Host,
		Port:        cfg.Port,
		ServiceName: cfg.ServiceName,
		SID:         cfg.SID,
		Credentials: Credentials{
			User:     cfg.User,
			Password: cfg.Password,
		},
		Timeouts: Timeouts{
			ConnectTimeout:          cfg.ConnectTimeout,
			TransportConnectTimeout: cfg.TransportConnectTimeout,
		},
	}
}

// newRACFromConfig creates a RACConfig from the generic Config structure.
// It converts NodeConfig entries to Node entries and extracts RAC-specific configuration
// such as load balancing, failover settings, retry count, and retry delay.
func newRACFromConfig(cfg Config) *RACConfig {
	nodes := make([]Node, len(cfg.Nodes))
	for i, n := range cfg.Nodes {
		nodes[i] = Node{
			Host:     n.Host,
			Port:     n.Port,
			Protocol: n.Protocol,
		}
	}

	return &RACConfig{
		Nodes:       nodes,
		ServiceName: cfg.ServiceName,
		Credentials: Credentials{
			User:     cfg.User,
			Password: cfg.Password,
		},
		Timeouts: Timeouts{
			ConnectTimeout:          cfg.ConnectTimeout,
			TransportConnectTimeout: cfg.TransportConnectTimeout,
		},
		LoadBalance: cfg.LoadBalance,
		Failover:    cfg.Failover,
		RetryCount:  cfg.RetryCount,
		RetryDelay:  cfg.RetryDelay,
	}
}

// newDataGuardFromConfig creates a DataGuardConfig from the generic Config structure.
// It extracts the primary node configuration, standby nodes, and DataGuard-specific settings
// such as failover mode, failover retries, and failover delay.
func newDataGuardFromConfig(cfg Config) *DataGuardConfig {
	standbys := make([]Node, len(cfg.Standbys))
	for i, n := range cfg.Standbys {
		standbys[i] = Node{
			Host:     n.Host,
			Port:     n.Port,
			Protocol: n.Protocol,
		}
	}

	return &DataGuardConfig{
		Primary: Node{
			Host:     cfg.Primary.Host,
			Port:     cfg.Primary.Port,
			Protocol: cfg.Primary.Protocol,
		},
		Standbys:    standbys,
		ServiceName: cfg.ServiceName,
		Credentials: Credentials{
			User:     cfg.User,
			Password: cfg.Password,
		},
		Timeouts: Timeouts{
			ConnectTimeout:          cfg.ConnectTimeout,
			TransportConnectTimeout: cfg.TransportConnectTimeout,
		},
		FailoverMode:    cfg.FailoverMode,
		FailoverRetries: cfg.FailoverRetries,
		FailoverDelay:   cfg.FailoverDelay,
	}
}
