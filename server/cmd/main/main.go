package main

import (
	"log"
	"net/http"

	"github.com/fanatic/robot-game/server"
)

func main() {
	s, err := server.NewState("my.db")
	if err != nil {
		log.Fatal(err)
	}

	r, err := server.New(s)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8000", r))
}
