package main

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

func readLoop(bufrw *bufio.ReadWriter, frameChan chan frame) {
	defer close(frameChan)

	for {
		fr, err := frameParser(bufrw)
		if err != nil {
			log.Error().Err(err).Msg("couldn't parse frame")
			return
		}
		// log.Debug().Interface("frame", fr).Str("payload", fr.payload).Msg("Received payload from client")
		switch fr.opcode {
		case OpcodePing:
			frameChan <- frame{
				payload: fr.payload,
				opcode:  OpcodeText,
			}
		case OpcodeText:
			frameChan <- frame{
				payload: "received message well, this is a text from server",
				opcode:  OpcodeText,
			}
		case OpcodeConnectionClose:
			frameChan <- frame{
				payload:    fr.payload,
				opcode:     OpcodeConnectionClose,
				statusCode: fr.statusCode,
			}
			return
		}
	}
}
