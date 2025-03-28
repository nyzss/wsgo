package main

import (
	"bufio"
	"crypto/sha1"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

var WEBSOCKET_MAGIC_GUID string = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func cleanup() {
	log.Info().Msg("cleaning up and exiting..")
}

func initSig() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()
}

func main() {
	initLogger()
	initSig()

	port := ":8080"

	http.HandleFunc("/", handleConnection)
	log.Info().Str("port", port).Msg("Serving on localhost")
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal().Str("port", port).Msg("couldn't listen and serve, current port")
	}
}

func upgradeConnection(bufrw *bufio.ReadWriter, clientKey string) error {
	accept := clientKey + WEBSOCKET_MAGIC_GUID
	h := sha1.Sum([]byte(accept))

	encoded := b64.StdEncoding.EncodeToString(h[:])

	log.Debug().Str("encoded key", encoded).Msg("")

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

func handleConnection(w http.ResponseWriter, r *http.Request) {
	connection := r.Header.Get("Connection")
	upgrade := r.Header.Get("Upgrade")
	clientKey := r.Header.Get("Sec-WebSocket-Key")
	version := r.Header.Get("Sec-WebSocket-Version")

	log.Info().Str("connection", connection).Str("upgrade", upgrade).Msg("")
	log.Info().Str("websocket_version", version).Str("websocket_key", clientKey).Msg("")

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

		frameChan := make(chan frame)

		go writeLoop(bufrw, frameChan)
		readLoop(bufrw, frameChan)

	} else {
		log.Warn().Msg("normal http request")
		// todo: remove after testing hijacking
		log.Fatal().Msg("testing hijacking")
	}
}
