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
func frameParser(bufrw *bufio.ReadWriter) (string, error) {
	readSize := 4096           // default reading and buffer size
	maxReadSize := 1024 * 1024 // max readsize
	frame, err := readChunk(bufrw, readSize)

	if err != nil {
		return "", err
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
	log.Debug().Int("header_size", hIndex).Msg("")
	for i := range payloadLen {
		// condition added in case we are at an index bigger than frame buffer (>= readSize in this case)
		// doubling size of frame each time to do less read() calls on the socket
		// (might want to check which is better, more read() calls or smaller buffer size)
		if len(frame) <= hIndex+j {
			if readSize <= maxReadSize { // continue expending readsize until we hit the 1mb
				readSize *= 2
			}
			log.Debug().Int("old_size", readSize/2).Int("new_size", readSize).Msg("doubling size of read()")
			frame, err = readChunk(bufrw, readSize)
			if err != nil {
				return "", err
			}
			j = 0
			hIndex = 0
		}
		unmaskedPayload[i] = frame[hIndex+int(j)] ^ maskingKey[i%4]
		j++
	}

	return string(unmaskedPayload), nil
}

func frameBuilder(payload string) []byte {

	l := len(payload)
	var numBytes int
	if l <= 65535 {
		numBytes = 2
	} else {
		numBytes = 8
	}

	// allocating for size of first 2 headers + extra payload length header(optional) + payload length
	buffer := make([]byte, 2+numBytes+l)
	var hIndex int

	buffer[hIndex] = 0x81 // 129
	hIndex++

	if l <= 125 {
		buffer[hIndex] = byte(l)
		hIndex++
	} else {
		hIndex++

		for i := numBytes - 1; i >= 0; i-- {
			buffer[hIndex+i] = byte(l & 0xFF)
			l >>= 8
		}

		hIndex += numBytes
	}

	for i := range payload {
		buffer[hIndex+i] = payload[i]
	}

	return buffer
}
