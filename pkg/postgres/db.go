package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Setup(cfg Config) (*sql.DB, error) {
	if err := CreateDatabase(cfg); err != nil {
		return nil, err
	}

	if cfg.PathOfMigration != "" {
		if err := RunMigrations(cfg); err != nil {
			return nil, err
		}
	}

	conn, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	))

	if err != nil {
		return nil, err
	}

	conn.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)

	return conn, err
}

// CreateDatabase creates a new database if it doesn't exist.
func CreateDatabase(cfg Config) error {
	conn, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.SSLMode,
	))

	if err != nil {
		return err
	}

	defer conn.Close()

	// Check if the database exists
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s');", cfg.DBName)
	err = conn.QueryRow(query).Scan(&exists)
	if err != nil {
		return err
	}

	// Create the database if it does not exist
	if !exists {
		_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE %s;", cfg.DBName))
	}
	return err
}

func RunMigrations(cfg Config) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.PathOfMigration),
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode),
	)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
