package main

import (
	"log"
	"net/http"

	"github.com/fanatic/robot-game/server"
)

func main() {
	r, err := server.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8000", r))
}
