package main

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

func readLoop(bufrw *bufio.ReadWriter, frameChan chan frame, stopChan chan struct{}) {
	defer close(frameChan)

	for {
		select {
		case <-stopChan:
			return
		default:
			fr, err := frameParser(bufrw)
			log.Info().Int("connection_id", 1).Msg("received message on read loop")

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
			case OpcodeText, OpcodeBinary:
				frameChan <- frame{
					// payload: "received message well, this is a text from server",
					payload: fr.payload,
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
}
