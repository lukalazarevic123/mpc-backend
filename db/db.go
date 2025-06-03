package db

import (
	"context"
	"errors"
	"fmt"
	"mpc-backend/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func NewMasterDb(config config.DbConfig) (*pgxpool.Pool, error) {
	// Create database connection to the master db
	masterDb, err := newDatabaseConnection(config)
	if err != nil {
		return nil, fmt.Errorf("could not connect to master database: %v", err)
	}

	// Run migrations on master db
	err = runMigrations(config)
	if err != nil {
		return nil, fmt.Errorf("could not run migrationas on master database: %v", err)
	}

	return masterDb, nil
}

func newDatabaseConnection(conf config.DbConfig) (*pgxpool.Pool, error) {
	log.Info().Str("conn", fmt.Sprintf("postgres://%s:*****@%s:%d/%s", conf.Username, conf.Host, conf.Port, conf.Database)).Msg("Connecting to database")
	db, err := pgxpool.New(context.Background(), conf.ConnectionString("postgres"))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(conf config.DbConfig) error {
	databaseConnectionString := conf.ConnectionString("postgres")
	m, err := migrate.New("file://./db/migrations", databaseConnectionString)

	if err != nil {
		return fmt.Errorf("failed to instantiate the driver: %w", err)
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Debug().Msg("Migrations up to date")
		} else {
			return fmt.Errorf("failed to migrate to latest: %w", err)
		}
	}

	return nil
}
