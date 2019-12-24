package impl

import (
	"github.com/orderbynull/lottip/core"
	"sort"
	"time"
)

func getData() []core.PgsqlPacket {
	return []core.PgsqlPacket{
		{Id: 20, Query: "SELECT * FROM shop ORDER BY article", Timestamp: time.Now()},
		{Id: 19, Query: "INSERT INTO shop VALUES (1,'A',3.45),(1,'B',3.99),(2,'A',10.99),(3,'B',1.45), (3,'C',1.69),(3,'D',1.25),(4,'D',19.95)", Timestamp: time.Now()},
		{Id: 18, Query: "CREATE TABLE shop ( article INT UNSIGNED DEFAULT '0000' NOT NULL, dealer CHAR(20) DEFAULT '' NOT NULL, price DECIMAL(16,2) DEFAULT '0.00' NOT NULL, PRIMARY KEY(article, dealer))", Timestamp: time.Now()},
		{Id: 17, Query: "SELECT year,month,BIT_COUNT(BIT_OR(1day)) AS days FROM t1 GROUP BY year,month", Timestamp: time.Now(), Error: "ERROR 1146 (42S02): Table 'test.no_such_table' doesn't exist"},
		{Id: 16, Query: "SELECT * FROM shop ORDER BY article", Timestamp: time.Now()},
		{Id: 15, Query: "INSERT INTO shop VALUES (1,'A',3.45),(1,'B',3.99),(2,'A',10.99),(3,'B',1.45), (3,'C',1.69),(3,'D',1.25),(4,'D',19.95)", Timestamp: time.Now()},
		{Id: 14, Query: "CREATE TABLE shop ( article INT UNSIGNED DEFAULT '0000' NOT NULL, dealer CHAR(20) DEFAULT '' NOT NULL, price DECIMAL(16,2) DEFAULT '0.00' NOT NULL, PRIMARY KEY(article, dealer))", Timestamp: time.Now()},
		{Id: 13, Query: "SELECT year,month,BIT_COUNT(BIT_OR(1day)) AS days FROM t1 GROUP BY year,month", Timestamp: time.Now(), Error: "ERROR 1146 (42S02): Table 'test.no_such_table' doesn't exist"},
		{Id: 12, Query: "SELECT * FROM shop ORDER BY article", Timestamp: time.Now()},
		{Id: 11, Query: "INSERT INTO shop VALUES (1,'A',3.45),(1,'B',3.99),(2,'A',10.99),(3,'B',1.45), (3,'C',1.69),(3,'D',1.25),(4,'D',19.95)", Timestamp: time.Now()},
		{Id: 10, Query: "CREATE TABLE shop ( article INT UNSIGNED DEFAULT '0000' NOT NULL, dealer CHAR(20) DEFAULT '' NOT NULL, price DECIMAL(16,2) DEFAULT '0.00' NOT NULL, PRIMARY KEY(article, dealer))", Timestamp: time.Now()},
		{Id: 9, Query: "SELECT year,month,BIT_COUNT(BIT_OR(1day)) AS days FROM t1 GROUP BY year,month", Timestamp: time.Now(), Error: "ERROR 1146 (42S02): Table 'test.no_such_table' doesn't exist"},
		{Id: 8, Query: "SELECT * FROM shop ORDER BY article", Timestamp: time.Now()},
		{Id: 7, Query: "INSERT INTO shop VALUES (1,'A',3.45),(1,'B',3.99),(2,'A',10.99),(3,'B',1.45), (3,'C',1.69),(3,'D',1.25),(4,'D',19.95)", Timestamp: time.Now()},
		{Id: 6, Query: "CREATE TABLE shop ( article INT UNSIGNED DEFAULT '0000' NOT NULL, dealer CHAR(20) DEFAULT '' NOT NULL, price DECIMAL(16,2) DEFAULT '0.00' NOT NULL, PRIMARY KEY(article, dealer))", Timestamp: time.Now()},
		{Id: 5, Query: "SELECT year,month,BIT_COUNT(BIT_OR(1day)) AS days FROM t1 GROUP BY year,month", Timestamp: time.Now(), Error: "ERROR 1146 (42S02): Table 'test.no_such_table' doesn't exist"},
		{Id: 4, Query: "SELECT * FROM shop ORDER BY article", Timestamp: time.Now()},
		{Id: 3, Query: "INSERT INTO shop VALUES (1,'A',3.45),(1,'B',3.99),(2,'A',10.99),(3,'B',1.45), (3,'C',1.69),(3,'D',1.25),(4,'D',19.95)", Timestamp: time.Now()},
		{Id: 2, Query: "CREATE TABLE shop ( article INT UNSIGNED DEFAULT '0000' NOT NULL, dealer CHAR(20) DEFAULT '' NOT NULL, price DECIMAL(16,2) DEFAULT '0.00' NOT NULL, PRIMARY KEY(article, dealer))", Timestamp: time.Now()},
		{Id: 1, Query: "SELECT year,month,BIT_COUNT(BIT_OR(1day)) AS days FROM t1 GROUP BY year,month", Timestamp: time.Now(), Error: "ERROR 1146 (42S02): Table 'test.no_such_table' doesn't exist"},
	}
}

type MemoryPgsqlRepository struct {
}

func (m MemoryPgsqlRepository) GetAllByApp(appName string, showType string, showedLast int, limit int) []core.PgsqlPacket {
	var filtered []core.PgsqlPacket
	
	data := getData()
	if showType == "newer" {
		sort.Slice(data, func(i, j int) bool {
			return data[i].Id < data[j].Id
		})
	} else {
		sort.Slice(data, func(i, j int) bool {
			return data[i].Id > data[j].Id
		})
	}

	for _, packet := range data {
		if len(filtered) == limit {
			break
		}

		if showedLast > 0 {
			switch showType {
			case "newer":
				if packet.Id > showedLast {
					filtered = append(filtered, packet)
				}
			case "older":
				if packet.Id < showedLast {
					filtered = append(filtered, packet)
				}
			}
		} else {
			filtered = append(filtered, packet)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return data[i].Id > data[j].Id
	})

	return filtered
}
