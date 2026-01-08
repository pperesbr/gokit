// Package dsn provides a flexible factory system for creating database DSN (Data Source Name) builders.
// It supports multiple database drivers through a registry pattern, allowing dynamic configuration
// loading from YAML files or raw bytes. Each driver can register its own builder factory to handle
// driver-specific DSN construction.
package dsn
