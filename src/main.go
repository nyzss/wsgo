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

func handleConnection(con http.ResponseWriter, req *http.Request) {
	fmt.Println("REQUEST: ", req)

	fmt.Println("---------------")

	connection := req.Header.Get("Connection")
	upgrade := req.Header.Get("Upgrade")
	key := req.Header.Get("Sec-WebSocket-Key")
	version := req.Header.Get("Sec-WebSocket-Version")

	fmt.Println("CONNECTION: ", connection)
	fmt.Println("UPGRADE: ", upgrade)
	fmt.Println("WEBSOCKET_KEY", key)
	fmt.Println("WEBSOCKET_VERSION", version)

	if connection == "Upgrade" && upgrade == "websocket" {
		fmt.Println("success: WEBSOCKET CONNECTION ASKED")
	} else {
		fmt.Println("warning: normal http request")
	}

}
