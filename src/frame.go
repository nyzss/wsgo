package main

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

func readChunk(bufrw *bufio.ReadWriter, n int) (chunk []byte, err error) {
	frame := make([]byte, n)
	b, err := bufrw.Read(frame)
	if err != nil {
		return nil, err
	}
	frame = frame[:b]
	return frame, nil
}

// ? parsing frame according to RFC 6455 section 5.2
func frameParser(bufrw *bufio.ReadWriter) {
	readSize := 4096           // default reading and buffer size
	maxReadSize := 1024 * 1024 // max readsize
	frame, err := readChunk(bufrw, readSize)

	if err != nil {
		log.Error().Err(err).Msg("error reading string")
		return
	}

	hIndex := 0                  // header index / header size
	fin := frame[hIndex] & 128   // 0x80
	rsv := frame[hIndex] & 112   // 0x70
	opcode := frame[hIndex] & 15 // 0x0F

	log.Debug().
		Int("header_index", hIndex).
		Bool("fin", fin != 0).
		Int("rsv", int(rsv)).
		Int("opcode", int(opcode)).
		Msg("Frame header first byte parsed")

	hIndex++
	masked := frame[hIndex] & 128

	var payloadLen int64

	payloadLenBase := frame[hIndex] & 127
	hIndex++
	if payloadLenBase <= 125 {
		payloadLen = int64(payloadLenBase)
	} else if payloadLenBase == 126 {
		payloadLen = int64((int(frame[hIndex]) << 8) | int(frame[hIndex+1]))
		hIndex += 2
	} else if payloadLenBase == 127 {
		for i := range 8 {
			payloadLen = (payloadLen << 8) | int64(frame[hIndex+i])
		}
		hIndex += 8
	}

	log.Debug().
		Int("header_size", hIndex).
		Bool("masked", masked != 0).
		Int64("payload_length", payloadLen).
		Msg("Frame header parsed")

	var maskingKey [4]byte

	// TODO: if masked is 0 then we should return an error (rfc 6455 section 5.1)
	if masked != 0 {
		for i := range 4 {
			maskingKey[i] = frame[hIndex+i]
		}
		hIndex += 4
		log.Debug().
			Hex("masking_key", maskingKey[:]).
			Msg("Masking key parsed")
	}

	// * UNMASKING CLIENT PAYLOAD HERE

	var unmaskedPayload []byte = make([]byte, payloadLen)

	j := 0
	for i := range payloadLen {
		// condition added in case we are at an index bigger than frame buffer (>= readSize in this case)
		// doubling size of frame each time to do less read() calls on the socket
		// (might want to check which is better, more read() calls or smaller buffer size)
		if len(frame) <= hIndex+j {
			if readSize <= maxReadSize { // continue expending readsize until we hit the 1mb
				readSize *= 2
			}
			log.Info().Int("new_size", readSize).Msg("")
			frame, err = readChunk(bufrw, readSize)
			if err != nil {
				return
			}
			j = 0
			hIndex = 0
		}
		unmaskedPayload[i] = frame[hIndex+int(j)] ^ maskingKey[i%4]
		j++
	}

	log.Debug().Str("payload", string(unmaskedPayload)).Msg("Received payload from client")
}
