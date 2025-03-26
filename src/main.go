package main

import (
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

func upgradeConnection(clientKey string) string {
	accept := clientKey + WEBSOCKET_MAGIC_GUID
	h := sha1.Sum([]byte(accept))

	encoded := b64.StdEncoding.EncodeToString(h[:])

	log.Info().Str("encoded key", encoded).Msg("")

	// HTTP/1.1 101 Switching Protocols
	// Upgrade: websocket
	// Connection: Upgrade
	// Sec-WebSocket-Accept: clientKey + WEBSOCKET_MAGIC_GUID

	return encoded
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
		upgradeConnection(clientKey)

		log.Info().Msg("Writing to client")

		bufrw.WriteString("writing from hijacked http server, bbb123\n")
		bufrw.Flush()

		s, err := bufrw.ReadString('\n')

		if err != nil {
			log.Error().Err(err).Msg("error reading string")
			return
		}

		log.Info().Str("message", s).Msg("received message from client")

		fmt.Fprintf(bufrw, "You said: %q\nBye.\n", s)
		bufrw.Flush()
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
