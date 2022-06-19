package main

import (
	"github.com/rs/zerolog/log"
	"reflect"
)

const hex = "01234567890ABCDEF"

func LogInvalid(info *ConnectionInfo, entryType string, packet []byte) {
	event := log.Log().Str("client", info.ClientAddress).Int("clientPort", info.ClientPort).
		Str("server", info.ServerAddress).Int("serverPort", info.ServerPort).
		Str("user", info.User).Str("type", entryType)

	if info.QueryId > 0 {
		event.Int("queryId", info.QueryId)
	}

	event.Int("packetType", int(GetPacketType(packet)))

	//packetStr := make([]byte, len(packet)*3)
	//for i, b := range packet {
	//	packetStr[i] = hex[b>>4]
	//	packetStr[i+1] = hex[b&0x0F]
	//	packetStr[i+2] = ' '
	//}
	//
	event.Bytes("packet", packet).Send()
}
func LogRequest(info *ConnectionInfo, entryType string, args ...interface{}) {
	if *logRequests || *logAll {
		sender := "client"
		doLogging(&sender, info, entryType, args)
	}
}

func LogResponse(info *ConnectionInfo, entryType string, args ...interface{}) {
	if *logResponses || *logAll {
		sender := "server"
		doLogging(&sender, info, entryType, args)
	}
}
func LogResponsePacket(info *ConnectionInfo, packet []byte) {
	if *logResponsePackets {
		sender := "server"
		args := make([]interface{}, 2)
		args[0] = "% x"
		args[1] = packet
		doLogging(&sender, info, "Response Packet", args)
	}
}

func LogOther(info *ConnectionInfo, entryType string, args ...interface{}) {
	if *logAll {
		doLogging(nil, info, entryType, args)
	}
}

func doLogging(sender *string, info *ConnectionInfo, entryType string, args []interface{}) {
	event := log.Log().Str("client", info.ClientAddress).Int("clientPort", info.ClientPort).
		Str("server", info.ServerAddress).Int("serverPort", info.ServerPort).
		Str("user", info.User).Str("type", entryType)

	if sender != nil {
		event.Str("sender", *sender)
	}

	if info.QueryId > 0 {
		event.Int("queryId", info.QueryId)
	}

	if len(args) > 0 {
		if len(args) == 1 {
			if reflect.TypeOf(args[0]).Kind() == reflect.String {
				event.Msg(args[0].(string))
			}
		} else {
			event.Msgf(args[0].(string), args[1:]...)
		}
	} else {
		event.Send()
	}
}
