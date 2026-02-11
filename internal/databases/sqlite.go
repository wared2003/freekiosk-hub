package databases

import (
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
)

func Open(path string) (*sqlx.DB, error) {

	db, err := sqlx.Connect("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	db.MustExec("PRAGMA journal_mode=WAL;")
	db.MustExec("PRAGMA synchronous=NORMAL;")
	db.MustExec("PRAGMA busy_timeout=5000;")
	db.MustExec("PRAGMA foreign_keys=ON;")

	return db, nil
}
