package db

import (
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

var Instance *goqu.Database

func InitializeDBInstance() error {
	db, err := util.GetDbConnection("")

	if err != nil {
		return err
	}

	Instance = goqu.New("postgres", db)

	return nil
}
