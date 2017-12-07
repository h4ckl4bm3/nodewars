package main

import (
	"log"
	"net/http"
	"nwmodel"
	"os"
)

func main() {
	certfile := os.Getenv("CERTFILE")
	keyfile := os.Getenv("KEYFILE")

	log.Println("Starting " + nwmodel.VersionTag + " server...")

	// Start Webserver
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", nwmodel.HandleConnections)

	err := http.ListenAndServeTLS(":443", certfile, keyfile, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
