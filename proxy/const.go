package proxy

const (
	responseEof         = 0xfe
	responseOk          = 0x00
	responsePrepareOk   = 0x00
	responseErr         = 0xff
	responseLocalinfile = 0xfb

	// MySQL field types constants
	fieldTypeString   = 0xfd
	fieldTypeLongLong = 0x08
	fieldTypeDouble   = 0x05

	// There is no code for Resultset in MySQL internal protocol
	// so it's defined here for convenience
	responseResultset = 0xbb

	// MySQL connection state constants
	connStateStarted  = 0xf4
	connStateFinished = 0xf5

	// Digits after comma
	doubleDecodePrecision = 6
)

const (
	comQuit byte = iota + 1
	comInitDB
	comQuery
	comFieldList
	comCreateDB
	comDropDB
	comRefresh
	comShutdown
	comStatistics
	comProcessInfo
	comConnect
	comProcessKill
	comDebug
	comPing
	comTime
	comDelayedInsert
	comChangeUser
	comBinlogDump
	comTableDump
	comConnectOut
	comRegisterSlave
	comStmtPrepare
	comStmtExecute
	comStmtSendLongData
	comStmtClose
	comStmtReset
	comSetOption
	comStmtFetch
)

// Capability flags
const (
	clientLongPassword uint32 = 1 << iota
	clientFoundRows
	clientLongFlag
	clientConnectWithDB
	clientNoSchema
	clientCompress
	clientODBC
	clientLocalFiles
	clientIgnoreSpace
	clientProtocol41
	clientInteractive
	clientSSL
	clientIgnoreSIGPIPE
	clientTransactions
	clientReserved
	clientSecureConnection
	clientMultiStatements
	clientMultiResults
	clientPSMultiResults
	clientPluginAuth
	clientConnectAttrs
	clientPluginAuthLenEncClientData
	clientCanHandleExpiredPasswords
	clientSessionTrack
	clientDeprecateEOF
)
