package proxy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strconv"
)

var errInvalidPacketLength = errors.New("Invalid packet length")
var errInvalidPacketType = errors.New("Invalid packet type")
var errFieldTypeNotImplementedYet = errors.New("Required field type not implemented yet")

// ComStmtPrepareOkResponse represents COM_STMT_PREPARE_OK response structure
type ComStmtPrepareOkResponse struct {

	// StatementID is an ID of prepared statement
	StatementID uint32

	// ParametersNum is num of prepared parameters
	ParametersNum uint16
}

// DecodeComStmtPrepareOkResponse decodes COM_STMT_PREPARE_OK response from MySQL server
// Basic packet structure shown below.
// For more details see https://mariadb.com/kb/en/mariadb/com_stmt_prepare/#COM_STMT_PREPARE_OK

// [0,1,2]:   int<3> PacketLength
// [3]: 	  int<1> PacketNumber
// [4]:       int<1> COM_STMT_PREPARE_OK (0x00)
// [5,6,7,8]: int<4> StatementID
// [9,10]:    int<2> NumberOfColumns
// [11,12]:   int<2> NumberOfParameters
// [13]:      string<1> <not used>
// [14,15]:   int<2> NumberOfWarnings
func DecodeComStmtPrepareOkResponse(packet []byte) (*ComStmtPrepareOkResponse, error) {

	// Min packet length = header(4 bytes) + command(1 byte) + statementID(4 bytes)
	// + number of columns (2 bytes) + number of parameters (2 bytes)
	// + <not used> (1 byte) + number of warnings (2 bytes)
	if len(packet) < 16 {
		return nil, errInvalidPacketLength
	}

	// Fifth byte is command
	if packet[4] != responsePrepareOk {
		return nil, errInvalidPacketType
	}

	statementID := binary.LittleEndian.Uint32(packet[5:9])
	parametersNum := binary.LittleEndian.Uint16(packet[11:13])

	return &ComStmtPrepareOkResponse{StatementID: statementID, ParametersNum: parametersNum}, nil
}

// ComStmtExecuteRequest represents COM_STMT_EXECUTE request structure
type ComStmtExecuteRequest struct {

	// StatementID is an ID of prepared statement
	StatementID uint32

	// PreparedParameters is a slice of prepared parameters
	PreparedParameters []PreparedParameter
}

// PreparedParameter structure represents single prepared parameter structure
// for COM_STMT_EXECUTE request
type PreparedParameter struct {

	// FieldType is type of prepared parameter
	// See https://mariadb.com/kb/en/mariadb/resultset/#field-types
	FieldType byte

	// Flags is unused
	Flag byte

	// Value is string presentation of any prepared parameter passed
	// with COM_STMT_EXECUTE request
	Value string
}

// DecodeComStmtExecuteRequest decodes COM_STMT_EXECUTE packet sent by MySQL client.
// Basic packet structure shown below.
// For more details see https://mariadb.com/kb/en/mariadb/com_stmt_execute/

// [0,1,2]:       int<3> PacketLength
// [3]: 	      int<1> PacketNumber
// [4]:           int<1> COM_STMT_EXECUTE (0x17)
// [5,6,7,8]:     int<4> StatementID
// [9]:           int<1> Flags
// [10,11,12, 13] int<4> IterationCount = 1
// 			  	  if (ParamCount > 0)
//			      {
// 				     byte<(ParamCount + 7) / 8> NullBitmap
// 				     byte<1>: SendTypeToServer = 0 or 1
// 				     if (SendTypeToServer)
//				     {
// 					    Foreach parameter
//					    {
// 						   byte<1>: FieldType
//						   byte<1>: ParameterFlag
//					    }
//				     }
// 				    Foreach parameter
//				    {
// 					   byte<n> BinaryParameterValue
//				    }
//			     }
func DecodeComStmtExecuteRequest(packet []byte, paramsCount uint16) (*ComStmtExecuteRequest, error) {

	// Min packet length = header(4 bytes) + command(1 byte) + statementID(4 bytes)
	// + flags(1 byte) + iteration count(4 bytes)
	if len(packet) < 14 {
		return nil, errInvalidPacketLength
	}

	// Fifth byte is command
	if packet[4] != requestComStmtExecute {
		return nil, errInvalidPacketType
	}

	reader := bytes.NewReader(packet)

	// Skip to statementID position
	reader.Seek(5, io.SeekStart)

	// Read StatementID value
	statementIDBuf := make([]byte, 4)
	reader.Read(statementIDBuf)
	statementID := binary.LittleEndian.Uint32(statementIDBuf)

	// Skip to NullBitmap position
	reader.Seek(5, io.SeekCurrent)

	parameters := make([]PreparedParameter, paramsCount)

	if paramsCount > 0 {
		nullBitmapLength := int64((paramsCount + 7) / 8)

		// Skip to SendTypeToServer position
		reader.Seek(nullBitmapLength, io.SeekCurrent)

		// Read SendTypeToServer value
		sendTypeToServerBuf := make([]byte, 1)
		reader.Read(sendTypeToServerBuf)
		sendTypeToServer, _ := DecodeLenEncodedInteger(sendTypeToServerBuf)

		if sendTypeToServer == 1 {
			for index := range parameters {

				// Read parameter FieldType and ParameterFlag
				parameterMeta := make([]byte, 2)
				reader.Read(parameterMeta)

				parameters[index].FieldType = parameterMeta[0]
				parameters[index].Flag = parameterMeta[1]
			}
		}

		var stringValue string

		for index, parameter := range parameters {
			switch parameter.FieldType {

			// MYSQL_TYPE_VAR_STRING
			// It's length encoded string
			case fieldTypeString:
				// Read first byte of parameter value to know buffer length for whole value
				stringLengthBuf := make([]byte, 1)
				reader.Read(stringLengthBuf)

				stringLength, _ := DecodeLenEncodedInteger(stringLengthBuf)
				reader.UnreadByte()

				// Packets with 0 length parameter are also possible
				if stringLength > 0 {
					// Read whole length encoded string
					stringValueBuf := make([]byte, stringLength+1)
					reader.Read(stringValueBuf)

					stringValue, _ = DecodeLenEncodedString(stringValueBuf)
				}

			// MYSQL_TYPE_LONGLONG
			case fieldTypeLongLong:
				var bigIntValue int64
				binary.Read(reader, binary.LittleEndian, &bigIntValue)

				stringValue = strconv.FormatInt(bigIntValue, 10)

			// MYSQL_TYPE_DOUBLE
			case fieldTypeDouble:
				// Read 8 bytes required for float64
				doubleLengthBuf := make([]byte, 8)
				reader.Read(doubleLengthBuf)

				doubleBits := binary.LittleEndian.Uint64(doubleLengthBuf)
				doubleValue := math.Float64frombits(doubleBits)

				stringValue = strconv.FormatFloat(doubleValue, 'f', 6, 64)

			default:
				return nil, errFieldTypeNotImplementedYet
			}

			parameters[index].Value = stringValue
		}
	}

	return &ComStmtExecuteRequest{StatementID: statementID, PreparedParameters: parameters}, nil
}

// DecodeLenEncodedInteger returns length-encoded integer and it's offset
// For more details see https://mariadb.com/kb/en/mariadb/protocol-data-types/#length-encoded-integers
func DecodeLenEncodedInteger(b []byte) (value uint64, offset uint64) {
	if len(b) == 0 {
		value = 0
		offset = 0
	}

	switch b[0] {
	case 0xfb:
		value = 0
		offset = 1

	case 0xfc:
		value = uint64(b[1]) | uint64(b[2])<<8
		offset = 3

	case 0xfd:
		value = uint64(b[1]) | uint64(b[2])<<8 | uint64(b[3])<<16
		offset = 4

	case 0xfe:
		value = uint64(b[1]) | uint64(b[2])<<8 | uint64(b[3])<<16 |
			uint64(b[4])<<24 | uint64(b[5])<<32 | uint64(b[6])<<40 |
			uint64(b[7])<<48 | uint64(b[8])<<56
		offset = 9

	default:
		value = uint64(b[0])
		offset = 1
	}

	return value, offset
}

// DecodeLenEncodedString returns length-encoded string and it's length
// Length-encoded strings are prefixed by a length-encoded integer which describes
// the length of the string, followed by the string value.
// For more details see https://mariadb.com/kb/en/mariadb/protocol-data-types/#length-encoded-strings
func DecodeLenEncodedString(b []byte) (string, uint64) {
	strLen, offset := DecodeLenEncodedInteger(b)

	return string(b[offset : offset+strLen]), strLen
}
