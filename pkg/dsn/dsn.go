// Package dsn provides interfaces and implementations for building Data Source Name (DSN) strings
// for different database systems. A DSN is a connection string that contains the information
// needed to establish a connection to a database.
package dsn

// DSN is an interface that defines the contract for building database connection strings.
// Implementations of this interface should provide the logic to construct a valid DSN
// string for their specific database system (e.g., PostgreSQL, MySQL, etc.).
type DSN interface {
	// Build constructs and returns a valid DSN string for the database connection.
	// It validates the configuration parameters and returns an error if any required
	// field is missing or invalid.
	//
	// Returns:
	//   - string: The constructed DSN connection string
	//   - error: An error if validation fails or the DSN cannot be built
	Build() (string, error)
}
