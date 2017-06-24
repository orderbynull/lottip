package proxy

import (
	"bytes"
	"errors"
	"io"
)

var errInvalidPacketLength = errors.New("Invalid packet length")
var errInvalidPacketType = errors.New("Invalid packet type")

type ComStmtExecuteRequest struct {
	StatementID        uint64
	PreparedParameters []PreparedParameter
}

type PreparedParameter struct {
	FieldType byte
	Flag      byte
	Value     string
}

//[0,1,2]:  int<3> PacketLength
//[3]: 	    int<1> PacketNumber
//[4]:      int<1> COM_STMT_EXECUTE (0x17)
//[5,6,7]:  int<4> StatementID
//[8]:      int<1> Flags
//[9,10,11] int<4> IterationCount = 1
// 			if (ParamCount > 0)
//			{
// 				byte<(ParamCount + 7) / 8> NullBitmap
// 				byte<1>: SendTypeToServer = 0 or 1
// 				if (SendTypeToServer)
//				{
// 					Foreach parameter
//					{
// 						byte<1>: FieldType
//						byte<1>: ParameterFlag
//					}
//				}
// 				Foreach parameter
//				{
// 					byte<n> BinaryParameterValue
//				}
//			}
func DecodeComStmtExecuteRequest(packet []byte, paramsCount int) (*ComStmtExecuteRequest, error) {
	if len(packet) < 14 {
		return nil, errInvalidPacketLength
	}

	if packet[4] != requestComStmtExecute {
		return nil, errInvalidPacketType
	}

	reader := bytes.NewReader(packet)

	// Skip to statementID position
	reader.Seek(5, io.SeekStart)

	//TODO: проверить как в драйвере MySQL считывается это значение
	// Read StatementID value
	statementIDBuf := make([]byte, 4)
	reader.Read(statementIDBuf)
	statementID, _, _ := readLenEncodedInt(statementIDBuf)

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
		sendTypeToServer, _, _ := readLenEncodedInt(sendTypeToServerBuf)

		if sendTypeToServer == 1 {
			for index, _ := range parameters {

				// Read parameter FieldType and ParameterFlag
				parameterMeta := make([]byte, 2)
				reader.Read(parameterMeta)

				parameters[index].FieldType = parameterMeta[0]
				parameters[index].Flag = parameterMeta[1]
			}
		}

		for index, parameter := range parameters {
			switch parameter.FieldType {
			case fieldTypeString:
				// Read first byte of parameter value to know buffer length for whole value
				stringLengthBuf := make([]byte, 1)
				reader.Read(stringLengthBuf)
				stringLength, _, _ := readLenEncodedInt(stringLengthBuf)
				reader.UnreadByte()

				// Read whole length encoded string
				stringValueBuf := make([]byte, stringLength+1)
				reader.Read(stringValueBuf)
				_, stringValue := readLenEncodedString(stringValueBuf)

				parameters[index].Value = stringValue
			}
		}
	}

	return &ComStmtExecuteRequest{StatementID: statementID, PreparedParameters: parameters}, nil
}

func readLenEncodedInt(b []byte) (uint64, uint64, bool) {
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

func readLenEncodedString(b []byte) (uint64, string) {
	strLen, offset, _ := readLenEncodedInt(b)

	return strLen, string(b[offset : offset+strLen])
}
