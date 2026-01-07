package db

import "testing"

func TestNewDatabaseConfig_WithMySQLDriver(t *testing.T) {
	config, err := NewDatabaseConfig(
		"mysql",
		"localhost",
		"test",
		"password",
		"test",
		3306,
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Driver != "mysql" {
		t.Errorf("expected driver 'mysql', got '%s'", config.Driver)
	}

	if config.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", config.Host)
	}

	if config.Port != 3306 {
		t.Errorf("expected port 3306, got %d", config.Port)
	}

	if config.Database != "test" {
		t.Errorf("expected database 'test', got '%s'", config.Database)
	}

	if config.DSN() != "test:password@tcp(localhost:3306)/test" {
		t.Errorf("expected DSN 'test:password@tcp(localhost:3306)/test', got '%s'", config.DSN())
	}
}

func TestNewDatabaseConfig_WithOracleDriver(t *testing.T) {
	config, err := NewDatabaseConfig(
		"oracle",
		"localhost",
		"test",
		"password",
		"test",
		1521,
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Driver != "oracle" {
		t.Errorf("expected driver 'oracle', got '%s'", config.Driver)
	}

	if config.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", config.Host)
	}

	if config.Port != 1521 {
		t.Errorf("expected port 1521, got %d", config.Port)
	}

	if config.Database != "test" {
		t.Errorf("expected database 'test', got '%s'", config.Database)
	}

	if config.DSN() != "oracle://test:password@localhost:1521/test" {
		t.Errorf("expected DSN 'test/password@localhost:1521/test', got '%s'", config.DSN())
	}

	config.Database = "sid:test"
	if config.DSN() != "oracle://test:password@localhost:1521?sid=test" {
		t.Errorf("expected DSN 'sid:test', got '%s'", config.DSN())
	}
}

func TestNewDatabaseConfig_WithPostgresDriver(t *testing.T) {
	config, err := NewDatabaseConfig(
		"postgres",
		"localhost",
		"test",
		"password",
		"test",
		5432)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Driver != "postgres" {
		t.Errorf("expected driver 'postgres', got '%s'", config.Driver)
	}

	if config.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", config.Host)
	}

	if config.Port != 5432 {
		t.Errorf("expected port 5432, got %d", config.Port)
	}

	if config.Database != "test" {
		t.Errorf("expected database 'test', got '%s'", config.Database)
	}

	if config.DSN() != "postgres://test:password@localhost:5432/test?sslmode=disable" {
		t.Errorf("expected DSN 'postgres://test:password@localhost:5432/test?sslmode=disable', got '%s'", config.DSN())
	}
}

func TestNewDatabaseConfig_WithInvalidDriver(t *testing.T) {
	_, err := NewDatabaseConfig("invalid", "", "", "", "", 0)
	if err == nil {
		t.Fatal("expected error for invalid driver")
	}
}

func TestNewDatabaseConfig_MissingHost(t *testing.T) {
	_, err := NewDatabaseConfig("postgres", "", "", "", "", 0)
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestNewDatabaseConfig_MissingUser(t *testing.T) {
	_, err := NewDatabaseConfig("postgres", "", "", "", "", 0)
	if err == nil {
		t.Fatal("expected error for empty user")
	}
}

func TestNewDatabaseConfig_MissingPassword(t *testing.T) {
	_, err := NewDatabaseConfig("postgres", "", "", "", "", 0)
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestNewDatabaseConfig_MissingDatabase(t *testing.T) {
	_, err := NewDatabaseConfig("postgres", "", "", "", "", 0)
	if err == nil {
		t.Fatal("expected error for empty database")
	}
}

func TestNewDatabaseConfig_MissingPortMySQLDriver(t *testing.T) {
	config, err := NewDatabaseConfig(
		"mysql",
		"localhost",
		"test",
		"password",
		"test",
		0,
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Port != 3306 {
		t.Errorf("expected port 3306, got %d", config.Port)
	}
}

func TestNewDatabaseConfig_MissingPortPostgresDriver(t *testing.T) {
	config, err := NewDatabaseConfig(
		"postgres",
		"localhost",
		"test",
		"password",
		"test",
		0,
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Port != 5432 {
		t.Errorf("expected port 5432, got %d", config.Port)
	}
}

func TestNewDatabaseConfig_MissingPortOracleDriver(t *testing.T) {
	config, err := NewDatabaseConfig(
		"oracle",
		"localhost",
		"test",
		"password",
		"test",
		0,
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Port != 1521 {
		t.Errorf("expected port 1521, got %d", config.Port)
	}
}
