package main

import (
	"log"
	"net/http"

	"github.com/De-cROMPOS/pastebin/linksmaker/internal/connectorclient"
)

func main() {
	var c connectorclient.ConnectorClient
	log.Printf("initializing connections to db...")
	err := c.Init()
	if err != nil {
		log.Fatalf("smth went wrong: %v", err)
	}
	log.Printf("connected, starting serving")

	http.HandleFunc("/", c.HashHandler)
	http.ListenAndServe(":1234", nil)

}
