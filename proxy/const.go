package proxy

const (
	// MySQL internal protocol constants
	requestCmdQuery       = 0x03
	requestCmdStmtPrepare = 0x16
	requestCmdStmtExecute = 0x17
	requestCmdStmtClose   = 0x19
	responseEof           = 0xfe
	responseOk            = 0x00
	responsePrepareOk     = 0x00
	responseErr           = 0xff
	responseLocalinfile   = 0xfb

	// There is no code for Resultset in MySQL internal protocol
	// so it's defined here for convenient usage
	responseResultset = 0xbb

	// MySQL connection state constants
	connStateStarted  = 0xf4
	connStateFinished = 0xf5
)
