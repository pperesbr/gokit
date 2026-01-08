package dsn_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/pperesbr/gokit/pkg/dsn"
	"github.com/pperesbr/gokit/pkg/dsn/mysql"
	"github.com/pperesbr/gokit/pkg/dsn/oracle"
	"github.com/pperesbr/gokit/pkg/dsn/postgres"
	"github.com/testcontainers/testcontainers-go"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/sijms/go-ora/v2"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupIntegrationFactory() *dsn.Factory {
	f := dsn.NewFactory()
	f.Register("oracle", oracle.NewBuilder)
	f.Register("postgres", postgres.NewBuilder)
	f.Register("mysql", mysql.NewBuilder)
	return f
}

func TestFactory_Integration_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Get container host and port
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	// Build connection string using factory
	yaml := fmt.Sprintf(`
host: %s
port: %d
database: testdb
user: testuser
password: testpass
sslmode: disable
`, host, port.Int())

	factory := setupIntegrationFactory()
	builder, err := factory.BuildFromDriver("postgres", []byte(yaml))
	if err != nil {
		t.Fatalf("failed to build postgres config: %v", err)
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("failed to generate connection string: %v", err)
	}

	// Test connection
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Verify connection works with a simple query
	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}

	if result != 1 {
		t.Errorf("unexpected result: got %d, want 1", result)
	}
}

func TestFactory_Integration_PostgresAutoDetect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	// Test auto-detect with top-level driver key
	yaml := fmt.Sprintf(`
postgres:
  host: %s
  port: %d
  database: testdb
  user: testuser
  password: testpass
  sslmode: disable
`, host, port.Int())

	factory := setupIntegrationFactory()
	builder, err := factory.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to auto-detect postgres config: %v", err)
	}

	if builder.Driver() != "postgres" {
		t.Errorf("unexpected driver: got %s, want postgres", builder.Driver())
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("failed to generate connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
}

func TestFactory_Integration_MySQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Start MySQL container
	container, err := tcmysql.Run(ctx,
		"mysql:8",
		tcmysql.WithDatabase("testdb"),
		tcmysql.WithUsername("testuser"),
		tcmysql.WithPassword("testpass"),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("failed to start mysql container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	// Build connection string using factory
	yaml := fmt.Sprintf(`
host: %s
port: %d
database: testdb
user: testuser
password: testpass
`, host, port.Int())

	factory := setupIntegrationFactory()
	builder, err := factory.BuildFromDriver("mysql", []byte(yaml))
	if err != nil {
		t.Fatalf("failed to build mysql config: %v", err)
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("failed to generate connection string: %v", err)
	}

	// Test connection
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Verify connection works with a simple query
	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}

	if result != 1 {
		t.Errorf("unexpected result: got %d, want 1", result)
	}
}

func TestFactory_Integration_MySQLAutoDetect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Start MySQL container
	container, err := tcmysql.Run(ctx,
		"mysql:8",
		tcmysql.WithDatabase("testdb"),
		tcmysql.WithUsername("testuser"),
		tcmysql.WithPassword("testpass"),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("failed to start mysql container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	// Test auto-detect with top-level driver key
	yaml := fmt.Sprintf(`
mysql:
  host: %s
  port: %d
  database: testdb
  user: testuser
  password: testpass
`, host, port.Int())

	factory := setupIntegrationFactory()
	builder, err := factory.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to auto-detect mysql config: %v", err)
	}

	if builder.Driver() != "mysql" {
		t.Errorf("unexpected driver: got %s, want mysql", builder.Driver())
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("failed to generate connection string: %v", err)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
}

func TestFactory_Integration_OracleStandalone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Start Oracle Free container using GenericContainer (may take a few minutes on first run)
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gvenzl/oracle-free:23-slim-faststart",
			ExposedPorts: []string{"1521/tcp"},
			Env: map[string]string{
				"ORACLE_PASSWORD": "testpass",
			},
			WaitingFor: wait.ForLog("DATABASE IS READY TO USE!").
				WithStartupTimeout(5 * time.Minute),
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("failed to start oracle container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "1521")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	// Build connection string using factory
	// Oracle Free uses service name FREEPDB1 by default
	yaml := fmt.Sprintf(`
mode: standalone
host: %s
port: %d
service_name: FREEPDB1
user: system
password: testpass
`, host, port.Int())

	factory := setupIntegrationFactory()
	builder, err := factory.BuildFromDriver("oracle", []byte(yaml))
	if err != nil {
		t.Fatalf("failed to build oracle config: %v", err)
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("failed to generate connection string: %v", err)
	}

	// Test connection using go-ora
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Verify connection works with a simple query
	var result int
	if err := db.QueryRow("SELECT 1 FROM DUAL").Scan(&result); err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}

	if result != 1 {
		t.Errorf("unexpected result: got %d, want 1", result)
	}
}

func TestFactory_Integration_OracleAutoDetect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Start Oracle Free container using GenericContainer
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gvenzl/oracle-free:23-slim-faststart",
			ExposedPorts: []string{"1521/tcp"},
			Env: map[string]string{
				"ORACLE_PASSWORD": "testpass",
			},
			WaitingFor: wait.ForLog("DATABASE IS READY TO USE!").
				WithStartupTimeout(5 * time.Minute),
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("failed to start oracle container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "1521")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	// Test auto-detect with top-level driver key
	yaml := fmt.Sprintf(`
oracle:
  mode: standalone
  host: %s
  port: %d
  service_name: FREEPDB1
  user: system
  password: testpass
`, host, port.Int())

	factory := setupIntegrationFactory()
	builder, err := factory.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to auto-detect oracle config: %v", err)
	}

	if builder.Driver() != "oracle" {
		t.Errorf("unexpected driver: got %s, want oracle", builder.Driver())
	}

	connStr, err := builder.ConnectionString()
	if err != nil {
		t.Fatalf("failed to generate connection string: %v", err)
	}

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
}
