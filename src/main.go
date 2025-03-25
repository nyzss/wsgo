package main

import (
	"fmt"
	"log"
	"net/http"
)

var _ = http.MethodGet

func main() {
	port := ":8080"
	fmt.Println("Hello, World!")

	http.HandleFunc("/", handleConnection)
	fmt.Printf("success: Serving on localhost%s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("fatal: couldn't listen and serve, current port: ", port)
	}
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	fmt.Println("REQUEST: ", r)

	fmt.Println("---------------")

	connection := r.Header.Get("Connection")
	upgrade := r.Header.Get("Upgrade")
	key := r.Header.Get("Sec-WebSocket-Key")
	version := r.Header.Get("Sec-WebSocket-Version")

	fmt.Println("CONNECTION: ", connection)
	fmt.Println("UPGRADE: ", upgrade)
	fmt.Println("WEBSOCKET_KEY", key)
	fmt.Println("WEBSOCKET_VERSION", version)

	// todo: setting as true for now to test hijacking, remove later on
	if true || connection == "Upgrade" && upgrade == "websocket" {
		fmt.Println("success: WEBSOCKET CONNECTION ASKED")
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

		defer conn.Close()
		bufrw.WriteString("writing from hijacked http server, bbb123")
		bufrw.Flush()

		s, err := bufrw.ReadString('\n')

		if err != nil {
			log.Printf("fatal: error reading string: %v", err)
			return
		}
		fmt.Fprintf(bufrw, "You said: %q\nBye.\n", s)
		bufrw.Flush()
	} else {
		fmt.Println("warning: normal http request")
		// todo: remove after testing hijacking
		log.Fatal("testing hijacking: you shouldn't see this ")
	}

}
