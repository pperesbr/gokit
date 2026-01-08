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
