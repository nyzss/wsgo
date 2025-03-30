package main

import (
	"bufio"
	"sync"

	"github.com/rs/zerolog/log"
)

func writeLoop(bufrw *bufio.ReadWriter, frameChan chan frame, stopChan chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for received := range frameChan {
		log.Info().Int("connection_id", 1).Msg("received channel message on write loop")
		data, shouldClose := frameBuilder(received)

		n, err := bufrw.Write(data)
		if err != nil {
			log.Error().Err(err).Int("bytes_written", n).Msg("couldn't write to client")
			return
		}

		if err := bufrw.Flush(); err != nil {
			log.Error().Err(err).Msg("couldn't flush buffered data to client")
			return
		}

		if shouldClose {
			close(stopChan)
			return
		}
	}
}

// if debug {
// 	log.Debug().Bytes("data", data).Msg("")
// 	fr, err := simpleFrameParser(data)
// 	if err != nil {
// 		log.Error().Err(err).Msg("couldn't parse built frame")
// 		return
// 	}

// 	fmt.Println("raw_buffer", data)
// 	fmt.Println("FRAME", fr)
// }

// dbg to check for elapsed time in us
// start := time.Now()
// elapsed := time.Since(start)
// log.Printf("flush took %s\n", elapsed)
