package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	InRange   []Robot   `json:"robots_in_range"`

	At  string `json:"at"`
	Msg string `json:"msg"`
}

const (
	North = iota
	East
	South
	West
)

func main() {
	name := "JP"
	if len(os.Args) == 2 {
		name = os.Args[1]
	}
	resp, err := http.Post("http://localhost:8000/robots", "application/json", strings.NewReader(`{"name": "`+name+`"}`))
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

	fmt.Printf("new robot: %+v\n", r)

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowUp:
				resp, err := http.Post("http://localhost:8000/robots/"+r.ID+"/move", "application/json", nil)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("move:       %s\n", Str(resp.Body))
			case termbox.KeyArrowLeft:
				resp, err := http.Post("http://localhost:8000/robots/"+r.ID+"/turn", "application/json", strings.NewReader(`{"direction":true}`))
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("turn-left:  %s\n", Str(resp.Body))
			case termbox.KeyArrowRight:
				resp, err := http.Post("http://localhost:8000/robots/"+r.ID+"/turn", "application/json", strings.NewReader(`{"direction":false}`))
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("turn-right: %s\n", Str(resp.Body))
			case termbox.KeySpace:
				resp, err := http.Post("http://localhost:8000/robots/"+r.ID+"/attack", "application/json", nil)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("attack:     %s\n", Str(resp.Body))
			case termbox.KeyEsc:
				break loop
			case 'q':
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

func Str(b io.Reader) string {
	var r Robot
	if err := json.NewDecoder(b).Decode(&r); err != nil {
		return err.Error()
	}
	if r.At == "error" {
		return fmt.Sprintf("error=%s", r.Msg)
	}
	return fmt.Sprintf("name=%s x=%d y=%d direction=%d score=%d dead=%t in-range=%v", r.Name, r.X, r.Y, r.Direction, r.Score, r.Dead, r.InRange)
}
