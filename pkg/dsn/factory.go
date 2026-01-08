// Package dsn provides a flexible factory system for creating database DSN (Data Source Name) builders.
// It supports multiple database drivers through a registry pattern, allowing dynamic configuration
// loading from YAML files or raw bytes. Each driver can register its own builder factory to handle
// driver-specific DSN construction.
package dsn

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// BuilderFactory is a function type that creates a Builder from raw configuration data.
// It takes a byte slice containing driver-specific configuration (typically YAML) and returns
// a Builder instance or an error if the configuration is invalid.
type BuilderFactory func(data []byte) (Builder, error)

// Factory manages the registration and creation of DSN builders for different database drivers.
// It maintains a registry of builder factories indexed by driver name and provides methods
// to load configurations from YAML files or byte arrays.
type Factory struct {
	// builders maps driver names to their corresponding BuilderFactory functions
	builders map[string]BuilderFactory
}

// NewFactory creates and initializes a new Factory instance with an empty builder registry.
// Drivers must be registered using the Register method before they can be used to build DSNs.
func NewFactory() *Factory {
	return &Factory{
		builders: make(map[string]BuilderFactory),
	}
}

// Register adds a new builder factory for the specified driver to the factory's registry.
// The driver name should match the key used in YAML configuration files.
// If a factory already exists for the driver, it will be replaced.
func (f *Factory) Register(driver string, factory BuilderFactory) {
	f.builders[driver] = factory
}

// LoadFromYAML reads a YAML configuration file from the specified path and creates a Builder.
// The YAML file should contain a top-level key matching a registered driver name, with the
// driver-specific configuration nested underneath. Returns an error if the file cannot be read,
// parsed, or if no registered driver is found in the configuration.
func (f *Factory) LoadFromYAML(path string) (Builder, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return f.LoadFromBytes(data)
}

// LoadFromBytes parses YAML configuration data and creates a Builder using the appropriate factory.
// The YAML data should contain a top-level key matching a registered driver name. The method
// iterates through the top-level keys until it finds a registered driver, then delegates
// the driver-specific configuration to the corresponding BuilderFactory.
// Returns an error if the YAML cannot be parsed or if no registered driver is found.
func (f *Factory) LoadFromBytes(data []byte) (Builder, error) {
	var raw map[string]yaml.Node
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	for driver, node := range raw {
		if factory, ok := f.builders[driver]; ok {
			driverData, err := yaml.Marshal(&node)
			if err != nil {
				return nil, fmt.Errorf("failed to extract %s config: %w", driver, err)
			}
			return factory(driverData)
		}
	}

	return nil, fmt.Errorf("no supported driver found in config")
}

// BuildFromDriver creates a Builder for the specified driver using the provided configuration data.
// Unlike LoadFromYAML and LoadFromBytes, this method expects the configuration data to contain
// only the driver-specific configuration without a top-level driver key.
// Returns an error if the driver is not registered or if the factory fails to create the builder.
func (f *Factory) BuildFromDriver(driver string, data []byte) (Builder, error) {
	factory, ok := f.builders[driver]
	if !ok {
		return nil, fmt.Errorf("unknown driver: %s", driver)
	}
	return factory(data)
}
