package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDatabase struct {
	container *postgres.PostgresContainer
	DB        *sql.DB
	ConnStr   string
}

func SetupTestDatabase(ctx context.Context) (*TestDatabase, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("tstdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		// postgres.WithInitScripts("../../migrations"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to start postgres: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return &TestDatabase{
		container: pgContainer,
		DB:        db,
		ConnStr:   connStr,
	}, nil
}

func (td *TestDatabase) Terminate(ctx context.Context) error {
	if td.DB != nil {
		td.DB.Close()
	}
	if td.container != nil {
		return td.container.Terminate(ctx)
	}
	return nil
}

func (td *TestDatabase) CleanDB(ctx context.Context) error {
	tables := []string{
		"public_links",
		"file_versions",
		"files",
		"sessions",
		"magic_links",
		"users",
		"events",
	}

	for _, table := range tables {
		_, err := td.DB.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Printf("Warning: failed to truncate %s: %v", table, err)
		}
	}
	return nil
}

func (td *TestDatabase) RunMigrations(migrationsPath string) error {
	driver, err := migratePostgres.WithInstance(td.DB, &migratePostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
