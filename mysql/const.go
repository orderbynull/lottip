package mysql

//ResponseOkPacket is sent from the server to the client to signal successful completion of a command.
//As of MySQL 5.7.5, OK packes are also used to indicate EOF, and EOF packets are deprecated.
const ResponseOkPacket byte = 0x00

const (
	//ComSleep is internal server command
	ComSleep byte = iota

	//ComQuit tells the server that the client wants to close the connection
	ComQuit

	//ComInitDB changes the default schema of the connection
	ComInitDB

	//ComQuery is used to send the server a text-based query that is executed immediately
	ComQuery

	//ComFieldList gets the column definitions of a table
	//As of MySQL 5.7.11, ComFieldList is deprecated and will be removed in a future version of MySQL
	ComFieldList

	//ComCreateDb creates a schema
	ComCreateDb

	//ComDropDB drops a schema
	ComDropDB

	//ComRefresh is low-level version of several FLUSH ... and RESET ... statements
	//As of MySQL 5.7.11, ComRefresh is deprecated and will be removed in a future version of MySQL
	ComRefresh

	//ComShutdown is used to shut down the MySQL server
	//As of MySQL 5.7.9, ComShutdown is deprecated and will be removed in MySQL 8.0
	ComShutdown

	//ComStatistics gets a human readable string of internal statistics
	ComStatistics

	//ComProcessInfo gets a list of active threads
	//As of MySQL 5.7.11, ComProcessInfo is deprecated and will be removed in a future version of MySQL
	ComProcessInfo

	//ComConnect is an internal command in the server
	ComConnect

	//ComProcessKill asks the server to terminate a connection
	//As of MySQL 5.7.11, ComProcessKill is deprecated and will be removed in a future version of MySQL
	ComProcessKill

	//ComDebug triggers a dump on internal debug info to stdout of the mysql-server
	ComDebug

	//ComPing checks if the server is alive
	ComPing

	//ComTime is an internal command in the server
	ComTime

	//ComDelayedInsert is an internal command in the server
	ComDelayedInsert

	//ComChangeUser changes the user of the current connection and reset the connection state
	ComChangeUser

	ComBinlogDump
	ComTableDump
	ComConnectOut
	ComRegisterSlave
	ComStmtPrepare
	ComStmtExecute
	ComStmtSendLongData
	ComStmtClose
	ComStmtReset
	ComSetOption
	ComStmtFetch
)

//MaxPacketSize is max size allowed for packet
const MaxPacketSize = 1<<24 - 1
