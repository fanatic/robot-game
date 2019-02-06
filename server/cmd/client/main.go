package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

type Robot struct {
	ID        string    `json:"id,omitempty"` // also the secret
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name" storm:"unique"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	Color     string    `json:"color"`
	Direction int       `json:"direction"`
	Vision    int       `json:"vision"`
	Score     int       `json:"score"`
	Dead      bool      `json:"dead"`
}

const (
	North = iota
	East
	South
	West
)

func main() {
	resp, err := http.Post("http://localhost:8000/robots", "application/json", strings.NewReader(`{"name": "JP"}`))
	if err != nil {
		log.Fatal(err)
	}
	var r Robot
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Fatal(err)
	}

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowUp:
				if _, err := http.Post("http://localhost:8000/robots/"+r.ID+"/move", "application/json", nil); err != nil {
					log.Fatal(err)
				}
				fmt.Println("move")
			case termbox.KeyArrowLeft:
				if _, err := http.Post("http://localhost:8000/robots/"+r.ID+"/turn", "application/json", strings.NewReader(`{"direction":true}`)); err != nil {
					log.Fatal(err)
				}
				fmt.Println("turn-left")
			case termbox.KeyArrowRight:
				if _, err := http.Post("http://localhost:8000/robots/"+r.ID+"/turn", "application/json", strings.NewReader(`{"direction":false}`)); err != nil {
					log.Fatal(err)
				}
				fmt.Println("turn-right")
			case termbox.KeySpace:
				if _, err := http.Post("http://localhost:8000/robots/"+r.ID+"/attack", "application/json", nil); err != nil {
					log.Fatal(err)
				}
				fmt.Println("!")
			case termbox.KeyEsc:
				break loop
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}

	req, err := http.NewRequest("DELETE", "http://localhost:8000/robots/"+r.ID, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
}
