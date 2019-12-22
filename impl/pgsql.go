package impl

import (
	"github.com/orderbynull/lottip/app"
	"time"
)

type MemoryPgsqlRepository struct {

}

func (m MemoryPgsqlRepository) GetAllByApp(appName string) []app.PgsqlPacket {
	return []app.PgsqlPacket{
		{Query: "SELECT * FROM shop ORDER BY article", Timestamp: time.Now()},
		{Query: "INSERT INTO shop VALUES (1,'A',3.45),(1,'B',3.99),(2,'A',10.99),(3,'B',1.45), (3,'C',1.69),(3,'D',1.25),(4,'D',19.95)", Timestamp: time.Now()},
		{Query: "CREATE TABLE shop ( article INT UNSIGNED DEFAULT '0000' NOT NULL, dealer CHAR(20) DEFAULT '' NOT NULL, price DECIMAL(16,2) DEFAULT '0.00' NOT NULL, PRIMARY KEY(article, dealer))", Timestamp: time.Now()},
		{Query: "SELECT year,month,BIT_COUNT(BIT_OR(1day)) AS days FROM t1 GROUP BY year,month", Timestamp: time.Now(), Error: "ERROR 1146 (42S02): Table 'test.no_such_table' doesn't exist"},
	}
}

