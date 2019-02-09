package main

import (
	"log"
	"net/http"
	"os"

	"github.com/fanatic/robot-game/server"
)

func main() {
	g, err := server.NewGame("my.db")
	if err != nil {
		log.Fatal(err)
	}

	r, err := server.New(g)
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Fatal(http.ListenAndServe(":"+port, r))
}
