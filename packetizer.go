package main

func extractPacketsFromBuffer(ci *ConnectionInfo, packetFragment *[]byte, buffer []byte, processPacket func(packet []byte)) {
	//LogOther(ci, "extractPacketsFromBuffer() Entry", "% x", buffer)

	var packet []byte
	if len(*packetFragment) > 0 {
		packet = append(*packetFragment, buffer...)
		*packetFragment = []byte{}
	} else {
		packet = buffer
	}

	//LogOther(ci, "extractPacketsFromBuffer() Combined", "% x", packet)

	offset := uint32(0)
	bufferLen := uint32(len(packet))
	for {
		if bufferLen == offset {
			// Nothing else
			break
		} else if offset < bufferLen && bufferLen-offset >= 4 {
			packetSize := uint32(packet[offset+0]) | uint32(packet[offset+1])<<8 | uint32(packet[offset+2])<<16
			if bufferLen >= offset+packetSize+4 {
				temp := make([]byte, 3+1+packetSize)
				copy(temp, packet[offset:offset+3+1+packetSize])

				// Now process the packet based on the current state of the connection
				//seqNum := int(temp[3])

				//LogOther(ci, "extractPacketsFromBuffer() Packet", "% x", temp)
				processPacket(temp)

				//fsm.Fire(protocol.PacketReceived, temp)
				//if temp[4] == 0 {
				//	okPacket()
				//} else if temp[4] == mysqlproto.ERR_PACKET {
				//	errCode := int32(temp[5]) | int32(temp[6])<<8
				//	errorPacket(errCode, string(temp[7:]))
				//} else if temp[4] == mysqlproto.EOF_PACKET {
				//	eofPacket(seqNum)
				//} else {
				//	// Data packets
				//	dataPacket(temp)
				//}
				offset += packetSize + 4
				continue
			}
		}

		if bufferLen-offset > 0 {
			*packetFragment = make([]byte, bufferLen-offset)
			copy(*packetFragment, packet[offset:bufferLen])
		} else {
			// We are done
			*packetFragment = []byte{}
		}
		break
	}
}
