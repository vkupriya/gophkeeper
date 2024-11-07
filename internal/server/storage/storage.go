package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vkupriya/gophkeeper/internal/server/models"
)

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrSecretAlreadyExists = errors.New("secret already exists")
	ErrSecretNotFound      = errors.New("secret not found")
	ErrNoSecrets           = errors.New("no secrets")
)

type PostgresDB struct {
	pool *pgxpool.Pool
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	return &PostgresDB{
		pool: pool,
	}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
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

func (p *PostgresDB) UserAdd(c *models.Config, u models.User) error {
	db := p.pool
	var pgErr *pgconn.PgError
	ctx, cancel := context.WithTimeout(context.Background(), c.ContextTimeout)
	defer cancel()

	querySQL := "INSERT INTO users (userid, password) VALUES($1, $2)"

	_, err := db.Exec(ctx, querySQL, u.UserID, u.Password)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to insert user %s into Postgres DB: %w", u.UserID, err)
	}
	return nil
}

func (p *PostgresDB) UserGet(c *models.Config, userid string) (models.User, error) {
	db := p.pool
	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), c.ContextTimeout)
	defer cancel()

	row := db.QueryRow(ctx, "SELECT * FROM users WHERE userid=$1", userid)
	err := row.Scan(&user.UserID, &user.Password)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.User{}, ErrUserNotFound
	case err != nil:
		return models.User{}, fmt.Errorf("failed to query user in DB: %w", err)
	}

	return user, nil
}

func (p *PostgresDB) SecretAdd(c *models.Config, userid string, secret *models.Secret) error {
	db := p.pool
	var pgErr *pgconn.PgError
	ctx, cancel := context.WithTimeout(context.Background(), c.ContextTimeout)
	defer cancel()

	querySQL := "INSERT INTO secrets (userid, name, type, meta, data, version) VALUES($1, $2, $3, $4, $5, $6)"

	_, err := db.Exec(ctx, querySQL, userid, secret.Name, secret.Type, secret.Meta, secret.Data, 1)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrSecretAlreadyExists
		}
		return fmt.Errorf("failed to insert secret %s into Postgres DB: %w", secret.Name, err)
	}
	return nil
}

func (p *PostgresDB) SecretUpdate(c *models.Config, userid string, secret *models.Secret) error {
	db := p.pool
	ctx, cancel := context.WithTimeout(context.Background(), c.ContextTimeout)
	defer cancel()

	querySQL := "UPDATE secrets SET version = version + 1, meta=$1, data=$2 WHERE (userid=$3 AND name=$4)"

	_, err := db.Exec(ctx, querySQL, secret.Meta, secret.Data, userid, secret.Name)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrSecretNotFound
	case err != nil:
		return fmt.Errorf("failed to update secret %s in Postgres DB: %w", secret.Name, err)
	}
	return nil
}

func (p *PostgresDB) SecretGet(c *models.Config, userid string, name string) (*models.Secret, error) {
	db := p.pool
	var secret models.Secret
	ctx, cancel := context.WithTimeout(context.Background(), c.ContextTimeout)
	defer cancel()

	querySQL := "SELECT * FROM secrets WHERE userid=$1 AND name=$2"

	row := db.QueryRow(ctx, querySQL, userid, name)
	err := row.Scan(&secret.UserID, &secret.Name, &secret.Type, &secret.Meta, &secret.Data, &secret.Version)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return &models.Secret{}, ErrSecretNotFound
	case err != nil:
		return &models.Secret{}, fmt.Errorf("failed to query secret: %w", err)
	}
	return &secret, nil
}

func (p *PostgresDB) SecretDelete(c *models.Config, userid string, name string) error {
	db := p.pool
	ctx, cancel := context.WithTimeout(context.Background(), c.ContextTimeout)
	defer cancel()

	querySQL := "DELETE FROM secrets WHERE userid=$1 AND name=$2"

	_, err := db.Exec(ctx, querySQL, userid, name)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrSecretNotFound
	case err != nil:
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

func (p *PostgresDB) SecretList(c *models.Config, userid string) (*models.SecretList, error) {
	db := p.pool
	secrets := models.SecretList{}
	ctx, cancel := context.WithTimeout(context.Background(), c.ContextTimeout)
	defer cancel()

	querySQL := "SELECT * FROM secrets WHERE userid=$1"

	rows, err := db.Query(ctx, querySQL, userid)
	if err != nil {
		return nil, fmt.Errorf("error querying secrets db: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s models.SecretItem
		if err = rows.Scan(
			nil,
			&s.Name,
			&s.Type,
			nil,
			nil,
			&s.Version,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row in secrets table: %w", err)
		}
		secrets = append(secrets, s)
	}

	if len(secrets) == 0 {
		return nil, ErrNoSecrets
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan rows in secrets table: %w", err)
	}

	return &secrets, nil
}

func (p *PostgresDB) Close() {
	p.pool.Close()
}
