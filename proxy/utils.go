package proxy

import (
	"strings"
)

func getUseDatabaseValue(query string) string {
	var db = ""

	words := strings.Fields(query)
	if len(words) == 2 && strings.ToUpper(words[0]) == "USE" {
		db = words[1]
	}

	return db
}
