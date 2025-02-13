package migrations

import (
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*
var fs embed.FS

func PostgresMigrate(connStr string) error {
	d, err := iofs.New(fs, "sql")
	if err != nil {
		return err
	}

	mig, err := migrate.NewWithSourceInstance("iofs", d, connStr)
	if err != nil {
		return err
	}
	if err := mig.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	}
	_, err = mig.Close()
	return err
}
func PostgresMigrateDown(connStr string) error {
	d, err := iofs.New(fs, "sql")
	if err != nil {
		return err
	}

	mig, err := migrate.NewWithSourceInstance("iofs", d, connStr)
	if err != nil {
		return err
	}
	if err := mig.Down(); err != nil {
		return err
	}
	_, err = mig.Close()
	return err
}
