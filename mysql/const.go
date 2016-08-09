package mysql

//ComSleep is internal server command
const ComSleep byte = 0x00

//ComQuit tells the server that the client wants to close the connection
const ComQuit byte = 0x01

//ComInitDb changes the default schema of the connection
const ComInitDb byte = 0x02

//ComQuery is used to send the server a text-based query that is executed immediately
const ComQuery byte = 0x03

//ComFieldList gets the column definitions of a table
//As of MySQL 5.7.11, ComFieldList is deprecated and will be removed in a future version of MySQL
const ComFieldList byte = 0x04

//ComCreateDb creates a schema
const ComCreateDb byte = 0x05

//ComDropDb drops a schema
const ComDropDb byte = 0x06

//ComRefresh is low-level version of several FLUSH ... and RESET ... statements
//As of MySQL 5.7.11, ComRefresh is deprecated and will be removed in a future version of MySQL
const ComRefresh byte = 0x07

//ComShutdown is used to shut down the MySQL server
//As of MySQL 5.7.9, ComShutdown is deprecated and will be removed in MySQL 8.0
const ComShutdown byte = 0x08

//ComStatistics gets a human readable string of internal statistics
const ComStatistics byte = 0x09

//ComProcessInfo gets a list of active threads
//As of MySQL 5.7.11, ComProcessInfo is deprecated and will be removed in a future version of MySQL
const ComProcessInfo byte = 0x0a

//ComConnect is an internal command in the server
const ComConnect byte = 0x0b

//ComProcessKill asks the server to terminate a connection
//As of MySQL 5.7.11, ComProcessKill is deprecated and will be removed in a future version of MySQL
const ComProcessKill byte = 0x0c

//ComDebug triggers a dump on internal debug info to stdout of the mysql-server
const ComDebug byte = 0x0d

//ComPing checks if the server is alive
const ComPing byte = 0x0e

//ComTime is an internal command in the server
const ComTime byte = 0x0f

//ComDelayedInsert is an internal command in the server
const ComDelayedInsert byte = 0x10

//ComChangeUser changes the user of the current connection and reset the connection state
const ComChangeUser byte = 0x11

//ComResetConnection resets the session state
const ComResetConnection byte = 0x1f

//ComDaemon is an internal command in the server
const ComDaemon byte = 0x1d

//MaxPacketSize is max size allowed for packet
const MaxPacketSize = 1<<24 - 1
