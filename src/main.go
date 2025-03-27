package main

import (
	"bufio"
	"crypto/sha1"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var WEBSOCKET_MAGIC_GUID string = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

}

func upgradeConnection(bufrw *bufio.ReadWriter, clientKey string) error {
	accept := clientKey + WEBSOCKET_MAGIC_GUID
	h := sha1.Sum([]byte(accept))

	encoded := b64.StdEncoding.EncodeToString(h[:])

	log.Info().Str("encoded key", encoded).Msg("")

	response := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n\r\n",
		encoded,
	)

	_, err := bufrw.WriteString(response)
	if err != nil {
		return err
	}

	if err := bufrw.Flush(); err != nil {
		return err
	}

	return nil
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
func frameParser(bufrw *bufio.ReadWriter) {
	readSize := 8192 // default reading to 8192
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
		// condition added in case we are at an index bigger than frame buffer (>= 8192 in this case)
		// doubling size of frame each time to do less read() calls on the socket
		// (might want to check which is better, more read() calls or smaller buffer size)
		if len(frame) <= hIndex+j {
			readSize *= 2
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

func main() {
	initLogger()

	port := ":8080"

	http.HandleFunc("/", handleConnection)
	log.Info().Str("port", port).Msg("Serving on localhost")
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal().Str("port", port).Msg("couldn't listen and serve, current port")
	}
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("==== REQUEST DETAILS START ====")

	connection := r.Header.Get("Connection")
	upgrade := r.Header.Get("Upgrade")
	clientKey := r.Header.Get("Sec-WebSocket-Key")
	version := r.Header.Get("Sec-WebSocket-Version")

	log.Info().Str("value", connection).Msg("CONNECTION")
	log.Info().Str("value", upgrade).Msg("UPGRADE")
	log.Info().Str("value", clientKey).Msg("WEBSOCKET_KEY")
	log.Info().Str("value", version).Msg("WEBSOCKET_VERSION")
	log.Info().Msg("==== REQUEST DETAILS END ====")

	// todo: setting as true for now to test hijacking, remove later on
	if true || connection == "Upgrade" && upgrade == "websocket" {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "fatal: websocket doesn't support hijacking", http.StatusInternalServerError)
			return
		}

		conn, bufrw, err := hj.Hijack()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() {
			log.Info().Str("address", conn.RemoteAddr().String()).Msg("client disconnected")
			conn.Close()
		}()
		log.Info().Msg("client connected successfully")
		log.Info().Str("address", conn.RemoteAddr().String()).Msg("client address")

		if err := upgradeConnection(bufrw, clientKey); err != nil {
			log.Error().Err(err).Msg("couldn't upgrade connection")
			return
		}

		frameParser(bufrw)
		// frameParser(buffer)

		// log.Info().Interface("message", buffer).Msg("received message from client")
		// fmt.Println("message:", hex.EncodeToString(buffer))

		bufrw.Flush()

		// fmt.Fprintf(bufrw, "You said: %q\nBye.\n", s)
		// log.Info().Msg("Writing to client")

		// bufrw.WriteString("writing from hijacked http server, bbb123\n")
		// bufrw.Flush()

	} else {
		log.Warn().Msg("normal http request")
		// todo: remove after testing hijacking
		log.Fatal().Msg("testing hijacking")
	}
}

/*
GET / HTTP/1.1
Host: localhost:8080
Connection: upgrade
Upgrade: websocket

*/

// 81 fe 08 c0 9f 30 96 20 e8
