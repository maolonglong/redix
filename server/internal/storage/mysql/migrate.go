package mysql

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

func autoMigrate(db *sql.DB) error {
	s, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil
	}
	d, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", s, "mysql", d)
	if err != nil {
		return nil
	}
	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}
