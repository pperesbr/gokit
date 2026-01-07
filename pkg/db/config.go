package db

import (
	"fmt"
	"strings"
)

// Driver represents a type for defining database drivers as string constants.
type Driver string

const (
	Postgres Driver = "postgres"
	MySQL    Driver = "mysql"
	Oracle   Driver = "oracle"
)

// DatabaseConfig defines the configuration required to connect to a database, including a driver, credentials, and settings.
type DatabaseConfig struct {
	Driver    Driver            `yaml:"driver"`
	Host      string            `yaml:"host"`
	User      string            `yaml:"user"`
	Password  string            `yaml:"pass"`
	Database  string            `yaml:"database"`
	Port      int               `yaml:"port"`
	ExtraArgs map[string]string `yaml:"extraArgs"`
}

// NewDatabaseConfig creates and validates a new DatabaseConfig with the provided driver, host, user, password, database, and port.
func NewDatabaseConfig(driver, host, user, password, database string, port int) (*DatabaseConfig, error) {
	cfg := &DatabaseConfig{
		Driver:   Driver(driver),
		Host:     host,
		User:     user,
		Password: password,
		Database: database,
		Port:     port,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks the DatabaseConfig fields for required values and defaults the port based on the selected database driver.
func (c *DatabaseConfig) Validate() error {
	switch c.Driver {
	case MySQL:
		if c.Port == 0 {
			c.Port = 3306
		}
	case Postgres:
		if c.Port == 0 {
			c.Port = 5432
		}

	case Oracle:
		if c.Port == 0 {
			c.Port = 1521
		}
	default:
		return fmt.Errorf("invalid driver: %s", c.Driver)
	}

	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.User == "" {
		return fmt.Errorf("user is required")
	}

	if c.Database == "" {
		return fmt.Errorf("database is required")
	}

	return nil
}

// DSN generates and returns the Data Source Name (DSN) string based on the database driver and configuration provided.
func (c *DatabaseConfig) DSN() string {
	switch c.Driver {
	case MySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.User, c.Password, c.Host, c.Port, c.Database)
	case Postgres:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Database)
	case Oracle:
		if strings.HasPrefix(c.Database, "sid:") {
			sid := strings.TrimPrefix(c.Database, "sid:")
			return fmt.Sprintf("oracle://%s:%s@%s:%d?sid=%s",
				c.User, c.Password, c.Host, c.Port, sid)
		}
		return fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
			c.User, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}
