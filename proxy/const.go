package proxy

const (
	// MySQL internal protocol constants
	requestComQuery       = 0x03
	requestComShowFields  = 0x04
	requestComStmtPrepare = 0x16
	requestComStmtExecute = 0x17
	requestComStmtClose   = 0x19

	responseEof         = 0xfe
	responseOk          = 0x00
	responsePrepareOk   = 0x00
	responseErr         = 0xff
	responseLocalinfile = 0xfb

	// MySQL field types constants
	fieldTypeString = 0xfd
	fieldTypeLongLong = 0x08

	// Extended client capabilities
	capabilityDeprecateEof = 0x100

	// There is no code for Resultset in MySQL internal protocol
	// so it's defined here for convenience
	responseResultset = 0xbb

	// MySQL connection state constants
	connStateStarted  = 0xf4
	connStateFinished = 0xf5
)
