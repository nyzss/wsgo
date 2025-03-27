package main

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

func readLoop(bufrw *bufio.ReadWriter) {
	for {
		frame, err := frameParser(bufrw)
		if err != nil {
			log.Error().Err(err).Msg("couldn't parse frame")
			return
		}

		log.Debug().Interface("frame", frame).Str("payload", frame.payload).Msg("Received payload from client")

		switch frame.opcode {
		case OpcodePing:
			data := frameBuilder(frame.payload, OpcodeText, 0)
			n, err := bufrw.Write(data)
			if err != nil {
				log.Error().Err(err).Int("bytes_written", n).Msg("couldn't write to client")
				return
			}
		case OpcodeText:
			data := frameBuilder("received message well, this is a text from server", OpcodeText, 0)
			n, err := bufrw.Write(data)
			if err != nil {
				log.Error().Err(err).Int("bytes_written", n).Msg("couldn't write to client")
				return
			}
		case OpcodeConnectionClose:
			data := frameBuilder(frame.payload, OpcodeConnectionClose, uint16(frame.statusCode))
			n, err := bufrw.Write(data)
			if err != nil {
				log.Error().Err(err).Int("bytes_written", n).Msg("couldn't close connection")
			}
			bufrw.Flush()
			return
		}
		bufrw.Flush()
	}
}
