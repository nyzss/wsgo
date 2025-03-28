package main

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

func writeLoop(bufrw *bufio.ReadWriter, frameChan chan frame) {
	for {
		received := <-frameChan

		data := frameBuilder(received)
		n, err := bufrw.Write(data)
		if err != nil {
			log.Error().Err(err).Int("bytes_written", n).Msg("couldn't write to client")
			return
		}

		bufrw.Flush()
		if received.opcode == OpcodeConnectionClose {
			return
		}
	}
}
