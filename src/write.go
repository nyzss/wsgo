package main

import (
	"bufio"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

func writeLoop(bufrw *bufio.ReadWriter, frameChan chan frame, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for received := range frameChan {

		data := frameBuilder(received)

		if debug {
			log.Debug().Bytes("data", data).Msg("")
			fr, err := simpleFrameParser(data)
			if err != nil {
				log.Error().Err(err).Msg("couldn't parse built frame")
				return
			}

			fmt.Println("raw_buffer", data)
			fmt.Println("FRAME", fr)
		}

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
