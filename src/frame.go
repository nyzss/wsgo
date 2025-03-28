package main

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

/*
 *  %x0 denotes a continuation frame
 *  %x1 denotes a text frame
 *  %x2 denotes a binary frame
 *  %x3-7 are reserved for further non-control frames
 *  %x8 denotes a connection close
 *  %x9 denotes a ping
 *  %xA denotes a pong
 *  %xB-F are reserved for further control frames
 */
type Opcode byte

const (
	OpcodeContinuation    Opcode = 0
	OpcodeText            Opcode = 1
	OopcodeBinary         Opcode = 2
	OpcodeConnectionClose Opcode = 8
	OpcodePing            Opcode = 9
	OpcodePong            Opcode = 10
)

type StatusCode uint16

type frame struct {
	fin          byte
	opcode       Opcode
	payload      string
	statusCode   StatusCode
	headerLength byte
	// rsv     byte
}

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
func frameParser(bufrw *bufio.ReadWriter) (frame, error) {
	readSize := 4096           // default reading and buffer size
	maxReadSize := 1024 * 1024 // max readsize
	chunk, err := readChunk(bufrw, readSize)

	if err != nil {
		log.Error().Msg("initial read failed")
		return frame{}, err
	}

	hIndex := 0                  // header index / header size
	fin := chunk[hIndex] & 128   // 0x80
	rsv := chunk[hIndex] & 112   // 0x70
	opcode := chunk[hIndex] & 15 // 0x0F

	hIndex++
	masked := chunk[hIndex] & 128

	var payloadLen int64

	payloadLenBase := chunk[hIndex] & 127
	hIndex++
	if payloadLenBase <= 125 {
		payloadLen = int64(payloadLenBase)
	} else if payloadLenBase == 126 {
		payloadLen = int64((int(chunk[hIndex]) << 8) | int(chunk[hIndex+1]))
		hIndex += 2
	} else if payloadLenBase == 127 {
		for i := range 8 {
			payloadLen = (payloadLen << 8) | int64(chunk[hIndex+i])
		}
		hIndex += 8
	}

	log.Debug().
		Int("header_size", hIndex).
		Bool("fin", fin != 0).
		Int("rsv", int(rsv)).
		Int("opcode", int(opcode)).
		Bool("masked", masked != 0).
		Int64("payload_length", payloadLen).
		Msg("Frame header parsed")

	var maskingKey [4]byte

	// TODO: if masked is 0 then we should return an error (rfc 6455 section 5.1)
	if masked != 0 {
		for i := range 4 {
			maskingKey[i] = chunk[hIndex+i]
		}
		hIndex += 4
		log.Debug().
			Hex("masking_key", maskingKey[:]).
			Int("header_size", hIndex).
			Msg("Masking key parsed")
	}

	// * UNMASKING CLIENT PAYLOAD HERE

	var statusCode StatusCode

	if opcode == byte(OpcodeConnectionClose) {
		for i := range 2 {
			statusCode = (StatusCode(byte(statusCode)) << 8) | StatusCode(chunk[hIndex+i]^maskingKey[i%4])
		}
		hIndex += 2
		log.Debug().
			Int("status_code", int(statusCode)).
			Int("header_size", hIndex).
			Msg("Connection close frame status code parsed")
	}

	var unmaskedPayload []byte = make([]byte, payloadLen)

	j := 0
	isClose := opcode == byte(OpcodeConnectionClose)
	if (isClose && payloadLen > 2) || (!isClose && payloadLen > 0) {
		h := hIndex
		for i := range payloadLen {
			// condition added in case we are at an index bigger than frame buffer (>= readSize in this case)
			// doubling size of frame each time to do less read() calls on the socket
			// (might want to check which is better, more read() calls or smaller buffer size)
			if len(chunk) <= h+j {
				if readSize <= maxReadSize { // continue expending readsize until we hit the 1mb
					readSize *= 2
				}
				log.Debug().
					Int("old_size", readSize/2).
					Int("new_size", readSize).
					Int("max_size", maxReadSize).
					Msg("Increasing read buffer size")
				chunk, err = readChunk(bufrw, readSize)
				if err != nil {
					log.Error().Msg("loop read failed")
					return frame{}, err
				}
				j = 0
				h = 0
			}
			unmaskedPayload[i] = chunk[h+int(j)] ^ maskingKey[i%4]
			j++
		}
	}

	return frame{
		fin:          fin,
		opcode:       Opcode(opcode),
		payload:      string(unmaskedPayload),
		statusCode:   statusCode,
		headerLength: byte(hIndex),
	}, nil
}

// * rfc-6455 section 5.2 (https://datatracker.ietf.org/doc/html/rfc6455#section-5.2)
func frameBuilder(fr frame) []byte {
	l := len(fr.payload)
	var numBytes int
	// payload len 125, 126 or 127
	if l > 125 && l <= 65535 {
		numBytes = 2
	} else if l > 65535 {
		numBytes = 8
	}

	headerSize := 2 + numBytes
	if fr.opcode == OpcodeConnectionClose {
		headerSize += 2
	}

	// allocating for size of first 2 headers + extra payload length header(optional) + payload length
	buffer := make([]byte, headerSize+l)
	var hIndex int

	buffer[hIndex] = 0x80 + byte(fr.opcode) // 128 + opcode
	hIndex++

	if l <= 125 {
		buffer[hIndex] = byte(l)
		hIndex++
	} else {
		if numBytes == 2 {
			buffer[hIndex] = 126
		} else {
			buffer[hIndex] = 127
		}
		hIndex++

		for i := numBytes - 1; i >= 0; i-- {
			buffer[hIndex+i] = byte(l & 0xFF)
			l >>= 8
		}

		hIndex += numBytes
	}

	// adding status code before setting payload
	if fr.opcode == OpcodeConnectionClose {
		statusCode := fr.statusCode
		for i := 1; i >= 0; i-- {
			buffer[hIndex+i] = byte(statusCode & 0xFF)
			statusCode >>= 8
		}
		hIndex += 2
	}

	// adding payload to buffer
	for i := range fr.payload {
		buffer[hIndex+i] = fr.payload[i]
	}

	return buffer
}
