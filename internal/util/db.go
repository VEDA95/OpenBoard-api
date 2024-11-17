package util

import (
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
)

func GetDbConnection(url string) (*sql.DB, error) {
	var dbUrl string

	if len(url) == 0 {
		dbUrl = os.Getenv("DB_URL")

		if len(dbUrl) == 0 {
			return nil, errors.New("DB_URL environment variable not set")
		}

	} else {
		dbUrl = url
	}

	db, err := sql.Open("pgx", dbUrl)

	if err != nil {
		return nil, err
	}

	return db, nil
}
