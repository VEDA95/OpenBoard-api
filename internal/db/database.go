package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"regexp"
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

type DB struct {
	connection *sql.DB
	Dialect    *goqu.Database
}

var Instance DB

func InitializeDBInstance() error {
	dbUrl := os.Getenv("DB_URL")

	if len(dbUrl) == 0 {
		return errors.New("DB_URL environment variable not set")
	}

	db, err := sql.Open("pgx", dbUrl)

	if err != nil {
		return err
	}

	Instance = DB{
		connection: db,
		Dialect:    goqu.Dialect("postgres").DB(db),
	}

	return nil
}

func ExtractSQLQueryString(query interface{}) (string, []interface{}, error) {
	var output string
	var err error
	var args []interface{}

	switch query.(type) {
	case *goqu.SelectDataset:
		castedQuery := query.(*goqu.SelectDataset)
		sql, queryArgs, err2 := castedQuery.ToSQL()
		output = sql
		err = err2
		args = queryArgs
		break

	case *goqu.InsertDataset:
		castedQuery := query.(*goqu.InsertDataset)
		sql, queryArgs, err2 := castedQuery.ToSQL()
		output = sql
		err = err2
		args = queryArgs
		break

	case *goqu.UpdateDataset:
		castedQuery := query.(*goqu.UpdateDataset)
		sql, queryArgs, err2 := castedQuery.ToSQL()
		output = sql
		err = err2
		args = queryArgs
		break

	case *goqu.DeleteDataset:
		castedQuery := query.(*goqu.DeleteDataset)
		sql, queryArgs, err2 := castedQuery.ToSQL()
		output = sql
		err = err2
		args = queryArgs
		break

	case *goqu.TruncateDataset:
		castedQuery := query.(*goqu.TruncateDataset)
		sql, queryArgs, err2 := castedQuery.ToSQL()
		output = sql
		err = err2
		args = queryArgs
		break

	default:
		return "", nil, errors.New(fmt.Sprintf("The datatype of the query provided is invalid: %v", query))

	}

	if err != nil {
		return "", nil, err
	}

	return output, args, nil
}

func (db *DB) Close() error {
	return db.connection.Close()
}

func (db *DB) ExecQuery(query interface{}) (*QueryResult, error) {
	sqlQuery, args, err := ExtractSQLQueryString(query)

	if err != nil {
		return nil, err
	}

	switch query.(type) {
	case *goqu.SelectDataset:
		break

	case *goqu.InsertDataset, *goqu.UpdateDataset, *goqu.DeleteDataset, *goqu.TruncateDataset:
		returningMatch, _ := regexp.MatchString(`RETURNING\s?`, sqlQuery)

		if returningMatch {
			break
		}

		_, err := db.connection.Exec(sqlQuery, args...)

		if err != nil {
			return nil, err
		}

		return nil, nil

	default:
		return nil, errors.New(fmt.Sprintf("The datatype of the query provided is invalid: %v", query))

	}

	rows, err := db.connection.Query(sqlQuery, args...)

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	columnCount := len(cols)
	var output []ExtractedRow

	for rows.Next() {
		row := make(map[string]interface{})
		values := make([]interface{}, columnCount)
		pointers := make([]interface{}, columnCount)

		for index := range cols {
			pointers[index] = &values[index]
		}

		rows.Scan(pointers...)

		for index, value := range values {
			row[cols[index]] = value
		}

		output = append(output, row)
	}

	return &QueryResult{Columns: cols, Rows: output, Size: len(output)}, nil
}

func (db *DB) ExecSingleQuery(query interface{}, columns []string) (*SingleQueryResult, error) {
	sqlQuery, args, err := ExtractSQLQueryString(query)

	if err != nil {
		return nil, err
	}

	switch query.(type) {
	case *goqu.SelectDataset:
		break

	case *goqu.InsertDataset, *goqu.UpdateDataset, *goqu.DeleteDataset, *goqu.TruncateDataset:
		returningMatch, _ := regexp.MatchString(`RETURNING\s?`, sqlQuery)

		if returningMatch {
			break
		}

		_, err := db.connection.Exec(sqlQuery, args...)

		if err != nil {
			return nil, err
		}

		return nil, nil

	default:
		return nil, errors.New(fmt.Sprintf("The datatype of the query provided is invalid: %v", query))
	}

	row := db.connection.QueryRow(sqlQuery, args...)
	columnCount := len(columns)
	output := make(map[string]interface{})
	pointers := make([]interface{}, columnCount)
	values := make([]interface{}, columnCount)

	for index := range columns {
		pointers[index] = &values[index]
	}

	row.Scan(pointers...)

	for index, value := range values {
		output[columns[index]] = value
	}

	return &SingleQueryResult{Columns: columns, Row: output}, nil
}
