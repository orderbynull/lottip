package main

import (
	"database/sql"

	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

//...
func getQueryResults(database, query string, params []string, dsn string) ([]string, [][]string, error) {
	//isPrepared := true

	// Open database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	if len(database) > 0 {
		_, err = db.Exec(fmt.Sprintf("USE %s;", database))
		if err != nil {
			return nil, nil, err
		}
	}

	// Prepare params
	var interfaceSlice = make([]interface{}, len(params))
	for i, d := range params {
		interfaceSlice[i] = d
	}

	// Execute query
	rows, err := db.Query(query, interfaceSlice...)
	if err != nil {
		return nil, nil, err
	}

	// Get result columns
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	resultRows := [][]string{}

	// Fetch rows
	for rows.Next() {
		tableRow := make([]string, len(columns))

		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, nil, err
		}

		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			tableRow[i] = value
		}

		resultRows = append(resultRows, tableRow)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	return columns, resultRows, nil
}

func getUseDatabaseValue(query string) string {
	var db = ""

	words := strings.Fields(query)
	if len(words) == 2 && strings.ToUpper(words[0]) == "USE" {
		db = words[1]
	}

	return db
}
