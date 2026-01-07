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
	AuthMethods     []ssh.AuthMethod    `yaml:"-"` // <- mudou
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

	err := cfg.Validate()
	if err != nil {
		return nil, err
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

// Validate checks the SSHConfig fields for required values, sets defaults, and prepares authentication methods.
func (c *SSHConfig) Validate() error {
	if c.Port == 0 {
		c.Port = 22
	}

	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.User == "" {
		return fmt.Errorf("user is required")
	}

	if c.Password == "" && c.KeyFile == "" {
		return fmt.Errorf("password or keyFile is required")
	}

	if c.KeyFile != "" {
		key, err := os.ReadFile(c.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to read keyFile: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse keyFile: %w", err)
		}

		c.AuthMethods = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		c.AuthMethods = []ssh.AuthMethod{
			ssh.Password(c.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = c.Password
				}
				return answers, nil
			}),
		}
	}

	if c.KnownHostsFile != "" {
		hostKeyCallback, err := knownhosts.New(c.KnownHostsFile)
		if err != nil {
			return fmt.Errorf("failed to load known_hosts: %w", err)
		}
		c.HostKeyCallback = hostKeyCallback
	} else {
		c.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	return nil
}
