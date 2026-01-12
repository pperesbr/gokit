package oracle

import "errors"

var (
	// ErrOracleHostRequired is returned when the host parameter is missing from the DSN.
	ErrOracleHostRequired = errors.New("oracle: host is required")

	// ErrOracleUserRequired is returned when the user parameter is missing from the DSN.
	ErrOracleUserRequired = errors.New("oracle: user is required")

	// ErrOraclePasswordRequired is returned when the password parameter is missing from the DSN.
	ErrOraclePasswordRequired = errors.New("oracle: password is required")

	// ErrOracleServiceNameRequired is returned when the database/service name parameter is missing from the DSN.
	ErrOracleServiceNameRequired = errors.New("oracle: database is required")

	// ErrOraclePortInvalid is returned when the port parameter is outside the valid range of 1-65535.
	ErrOraclePortInvalid = errors.New("oracle: port must between 1-65535")

	// ErrOracleConnectTimeoutInvalid is returned when the connect_timeout parameter is negative.
	ErrOracleConnectTimeoutInvalid = errors.New("oracle: connect_timeout must be greater than or equal to 0")

	// ErrOracleTimeoutInvalid is returned when the timeout parameter is negative.
	ErrOracleTimeoutInvalid = errors.New("oracle: timeout must be greater than or equal to 0")
)
