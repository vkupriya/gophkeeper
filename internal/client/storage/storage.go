package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/vkupriya/gophkeeper/internal/client/models"
)

const ctxTimeoutDefault time.Duration = time.Second * 3

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrSecretAlreadyExists = errors.New("secret already exists")
	ErrSecretNotFound      = errors.New("secret not found")
	ErrNoSecrets           = errors.New("no secrets")
)

type SQLiteDB struct {
	DB *sql.DB
}

func NewSQLiteDB(dbpath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}
	return &SQLiteDB{DB: db}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func RunMigrations(s *SQLiteDB) error {
	sourceInstance, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}
	targetInstance, err := sqlite.WithInstance(s.DB, new(sqlite.Config))
	if err != nil {
		return fmt.Errorf("invalid target sqlite instance, %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceInstance, "sqlite3", targetInstance)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (s *SQLiteDB) SecretList() ([]*models.SecretItem, error) {
	db := s.DB
	secrets := make([]*models.SecretItem, 0)
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDefault)
	defer cancel()

	querySQL := "SELECT name, type, version FROM secrets"

	rows, err := db.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, fmt.Errorf("error querying secrets db: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("failed to close rows")
		}
	}()

	for rows.Next() {
		var s models.SecretItem
		if err = rows.Scan(
			&s.Name,
			&s.Type,
			&s.Version,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row in secrets table: %w", err)
		}
		secrets = append(secrets, &s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan rows in secrets table: %w", err)
	}

	return secrets, nil
}

func (s *SQLiteDB) SecretDelete(name string) error {
	db := s.DB
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDefault)
	defer cancel()

	querySQL := "DELETE FROM secrets WHERE name=?"

	_, err := db.ExecContext(ctx, querySQL, name)
	if err != nil {
		return fmt.Errorf("error deleting secret %s: %w", name, err)
	}

	return nil
}

func (s *SQLiteDB) SecretDeleteAll() error {
	db := s.DB
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDefault)
	defer cancel()

	querySQL := "DELETE FROM secrets"

	_, err := db.ExecContext(ctx, querySQL)
	if err != nil {
		return fmt.Errorf("error deleting all secrets: %w", err)
	}

	return nil
}

func (s *SQLiteDB) SecretAdd(secret *models.Secret) error {
	db := s.DB
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDefault)
	defer cancel()

	querySQL := "INSERT INTO secrets (name, type, meta, data, version) VALUES(?,?,?,?,?)"

	_, err := db.ExecContext(ctx, querySQL, secret.Name, secret.Type, secret.Meta, secret.Data, secret.Version)
	if err != nil {
		return fmt.Errorf("failed to insert secret %s into SQLiteDB: %w", secret.Name, err)
	}
	return nil
}

func (s *SQLiteDB) SecretUpdate(secret *models.Secret) error {
	db := s.DB
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDefault)
	defer cancel()

	querySQL := "UPDATE secrets SET type=?, meta=?, data=?, version=? WHERE name=?"

	_, err := db.ExecContext(ctx, querySQL, secret.Type, secret.Meta, secret.Data, secret.Version, secret.Name)
	if err != nil {
		return fmt.Errorf("failed to update secret %s in SQLiteDB: %w", secret.Name, err)
	}
	return nil
}

func (s *SQLiteDB) SecretGet(name string) (*models.Secret, error) {
	db := s.DB
	var secret models.Secret
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDefault)
	defer cancel()

	querySQL := "SELECT * FROM secrets WHERE name=?"

	row := db.QueryRowContext(ctx, querySQL, name)
	err := row.Scan(&secret.Name, &secret.Type, &secret.Meta, &secret.Data, &secret.Version)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return &models.Secret{}, ErrSecretNotFound
	case err != nil:
		return &models.Secret{}, fmt.Errorf("failed to query secret: %w", err)
	}
	return &secret, nil
}
