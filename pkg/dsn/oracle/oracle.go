package oracle

const (
	// DriverName is the name of the Oracle database driver.
	DriverName = "oracle"
	// DefaultPort is the default port number for Oracle database connections.
	DefaultPort = 1521
	// DefaultProtocol is the default network protocol used for Oracle database connections.
	DefaultProtocol = "TCP"
)

// Credentials holds the authentication information for Oracle database connections.
type Credentials struct {
	// User is the username for database authentication.
	User string
	// Password is the password for database authentication.
	Password string
}

// Timeouts defines timeout values for Oracle database connection operations.
type Timeouts struct {
	// ConnectTimeout is the maximum time in seconds to wait for a connection to be established.
	ConnectTimeout int
	// TransportConnectTimeout is the maximum time in seconds to wait for the transport layer connection.
	TransportConnectTimeout int
}

// Node represents a single Oracle database instance endpoint in a cluster configuration.
// It defines the connection parameters needed to establish a connection to a specific node.
type Node struct {
	// Host is the hostname or IP address of the Oracle database instance.
	Host string
	// Port is the port number on which the Oracle database instance is listening.
	// If not specified, DefaultPort (1521) will be used.
	Port int
	// Protocol specifies the network protocol used for the connection (e.g., TCP, TCPS).
	// If not specified, DefaultProtocol (TCP) will be used.
	Protocol string
}

// normalizeNode applies default values to a Node configuration.
// It sets the Port to DefaultPort (1521) if not specified and
// sets the Protocol to DefaultProtocol (TCP) if not specified.
// This ensures that all Node instances have valid connection parameters.
func normalizeNode(n Node) Node {
	if n.Port == 0 {
		n.Port = DefaultPort
	}
	if n.Protocol == "" {
		n.Protocol = DefaultProtocol
	}
	return n
}
