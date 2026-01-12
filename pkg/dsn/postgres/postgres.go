package postgres

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/pperesbr/gokit/pkg/dsn"
)

var (
	_             dsn.DSN = (*Config)(nil)
	validSSLModes         = map[string]struct{}{
		"disable":     {},
		"allow":       {},
		"prefer":      {},
		"require":     {},
		"verify-ca":   {},
		"verify-full": {},
	}

	ErrPostgresHostRequired          = errors.New("postgres: host is required")
	ErrPostgresUserRequired          = errors.New("postgres: user is required")
	ErrPostgresPasswordRequired      = errors.New("postgres: password is required")
	ErrPostgresDatabaseRequired      = errors.New("postgres: database is required")
	ErrPostgresInvalidPort           = errors.New("postgres: port must between 1-65535")
	ErrPostgresInvalidSSLMode        = errors.New("postgres: invalid sslmode value, valid values are: disable, allow, prefer, require, verify-ca, verify-full")
	ErrPostgresInvalidConnectTimeout = errors.New("postgres: connect_timeout must be greater than or equal to 0")
)

type Config struct {
	Host            string `yaml:"host"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Port            int    `yaml:"port"`
	SSLMode         string `yaml:"ssl_mode"`
	ApplicationName string `yaml:"application_name"`
	ConnectTimeout  *int   `yaml:"connection_timeout"`
	SearchPath      string `yaml:"search_path"`
	Timezone        string `yaml:"timezone"`
}

func (c *Config) Build() (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}

	var params []string
	if c.SSLMode != "" {
		params = append(params, fmt.Sprintf("sslmode=%s", c.SSLMode))
	}

	if c.ApplicationName != "" {
		params = append(params, fmt.Sprintf("application_name=%s", url.QueryEscape(c.ApplicationName)))
	}

	if c.ConnectTimeout != nil && *c.ConnectTimeout >= 0 {
		params = append(params, fmt.Sprintf("connect_timeout=%d", *c.ConnectTimeout))
	}

	if c.SearchPath != "" {
		params = append(params, fmt.Sprintf("search_path=%s", url.QueryEscape(c.SearchPath)))
	}

	if c.Timezone != "" {
		params = append(params, fmt.Sprintf("timezone=%s", url.QueryEscape(c.Timezone)))
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		c.Host,
		c.Port,
		c.Database,
	)

	if len(params) > 0 {
		dsn = dsn + "?" + strings.Join(params, "&")
	}

	return dsn, nil

}

func (c *Config) validate() error {
	if c.Host == "" {
		return ErrPostgresHostRequired
	}

	if c.User == "" {
		return ErrPostgresUserRequired
	}

	if c.Password == "" {
		return ErrPostgresPasswordRequired
	}

	if c.Database == "" {
		return ErrPostgresDatabaseRequired
	}

	if c.Port == 0 {
		c.Port = 5432
	}

	if c.Port < 0 || c.Port > 65535 {
		return ErrPostgresInvalidPort
	}

	if c.SSLMode != "" && !isValidSSLMode(c.SSLMode) {
		return ErrPostgresInvalidSSLMode
	}

	if c.ConnectTimeout != nil && *c.ConnectTimeout < 0 {
		return ErrPostgresInvalidConnectTimeout
	}

	return nil
}

func isValidSSLMode(mode string) bool {
	_, ok := validSSLModes[mode]
	return ok
}
