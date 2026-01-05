package tunnel

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHConfig represents the configuration for establishing an SSH connection, including authentication and host details.
type SSHConfig struct {
	User            string              `yaml:"user"`
	Password        string              `yaml:"password"`
	KeyFile         string              `yaml:"keyFile"`
	Host            string              `yaml:"host"`
	KnownHostsFile  string              `yaml:"knownHostsFile"`
	Port            int                 `yaml:"port"`
	AuthMethod      ssh.AuthMethod      `yaml:"-"`
	HostKeyCallback ssh.HostKeyCallback `yaml:"-"`
}

// NewSSHConfig creates and returns a new SSHConfig object with the specified parameters and performs required validations.
func NewSSHConfig(user, password, keyFile, host, knownHostsFile string, port int) (*SSHConfig, error) {
	cfg := &SSHConfig{
		User:           user,
		Password:       password,
		KeyFile:        keyFile,
		Host:           host,
		KnownHostsFile: knownHostsFile,
		Port:           port,
	}

	if cfg.Port == 0 {
		cfg.Port = 22
	}

	if cfg.Host == "" {
		return nil, fmt.Errorf("host is required")
	}

	if cfg.User == "" {
		return nil, fmt.Errorf("user is required")
	}

	if cfg.Password == "" && cfg.KeyFile == "" {
		return nil, fmt.Errorf("password or keyFile is required")
	}

	if cfg.KeyFile != "" {
		key, err := os.ReadFile(cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read keyFile: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse keyFile: %w", err)
		}

		cfg.AuthMethod = ssh.PublicKeys(signer)
	} else {
		cfg.AuthMethod = ssh.Password(cfg.Password)
	}

	if cfg.KnownHostsFile != "" {
		hostKeyCallback, err := knownhosts.New(cfg.KnownHostsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load known_hosts: %w", err)
		}
		cfg.HostKeyCallback = hostKeyCallback
	} else {
		cfg.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	return cfg, nil
}

// Addr returns the SSH host and port formatted as a string in the "host:port" format.
func (c *SSHConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsInsecure determines if the SSHConfig lacks a KnownHostsFile, implying an insecure host key verification strategy.
func (c *SSHConfig) IsInsecure() bool {
	return c.KnownHostsFile == ""
}
