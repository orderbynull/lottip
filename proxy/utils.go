package proxy

import "strings"

func lenDecInt(b []byte) (uint64, uint64, bool) {
	if len(b) == 0 {
		return 0, 0, true
	}

	switch b[0] {
	case 0xfb:
		return 0, 1, true
	case 0xfc:
		return uint64(b[1]) | uint64(b[2])<<8, 3, false
	case 0xfd:
		return uint64(b[1]) | uint64(b[2])<<8 | uint64(b[3])<<16, 4, false
	case 0xfe:
		return uint64(b[1]) | uint64(b[2])<<8 | uint64(b[3])<<16 |
			uint64(b[4])<<24 | uint64(b[5])<<32 | uint64(b[6])<<40 |
			uint64(b[7])<<48 | uint64(b[8])<<56, 9, false
	default:
		return uint64(b[0]), 1, false
	}
}

func haventYetDecidedFuncName(query string) string {
	var db = ""

	words := strings.Fields(query)
	if len(words) == 2 && strings.ToUpper(words[0]) == "USE" {
		db = words[1]
	}

	return db
}
