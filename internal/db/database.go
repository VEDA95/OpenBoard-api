package db

import (
	"database/sql"
	"errors"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
)

type ExtractedRow map[string]interface{}

type QueryResult struct {
	Columns []string
	Rows    []ExtractedRow
	Size    int
}

type SingleQueryResult struct {
	Columns []string
	Row     ExtractedRow
}

var Instance *goqu.Database

func InitializeDBInstance() error {
	dbUrl := os.Getenv("DB_URL")

	if len(dbUrl) == 0 {
		return errors.New("DB_URL environment variable not set")
	}

	db, err := sql.Open("pgx", dbUrl)

	if err != nil {
		return err
	}

	Instance = goqu.New("postgres", db)

	return nil
}
