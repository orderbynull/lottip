package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func getQueryResults(query string) ([]string, [][]string, error) {

	// Open database
	db, err := sql.Open("mysql", "root:root@/")
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	// Execute query
	rows, err := db.Query(query)
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
