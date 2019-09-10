package main

import (
	"github.com/scaleway/functions-runtime/server"
	"log"
)

func main() {
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
