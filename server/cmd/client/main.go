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

// APIEndpoint is HTTP endpoint of the robot game api (omit trailing slash)
const APIEndpoint = "http://localhost:8000"

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("Usage: client INITIALS")
		return
	}
	r := call("POST", "/robots", `{"name": "`+os.Args[1]+`"}`)

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)

	fmt.Printf("new robot: %+v\n", r)

	loop(r.ID)

	termbox.Close()

	call("DELETE", "/robots/"+r.ID, "")
}

func loop(id string) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowUp:
				fmt.Printf("move:       ")
				r := call("POST", "/robots/"+id+"/move", "")
				fmt.Println(r)

			case termbox.KeyArrowLeft:
				fmt.Printf("turn-left:  ")
				r := call("POST", "/robots/"+id+"/turn", `{"direction":true}`)
				fmt.Println(r)

			case termbox.KeyArrowRight:
				fmt.Printf("turn-right: ")
				r := call("POST", "/robots/"+id+"/turn", `{"direction":false}`)
				fmt.Println(r)

			case termbox.KeySpace:
				fmt.Printf("attack:     ")
				r := call("POST", "/robots/"+id+"/attack", "")
				fmt.Println(r)

			case termbox.KeyEsc:
				return
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

// call makes a request to the robot game API
func call(method, path, payload string) *Robot {
	var body io.Reader
	if payload != "" {
		body = strings.NewReader(payload)
	}
	req, err := http.NewRequest(method, APIEndpoint+path, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var r Robot
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Fatalln(err)
	}
	return &r
}

type Robot struct {
	ID        string    `json:"id"` // also the secret
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	Color     string    `json:"color"`
	Direction int       `json:"direction"`
	Vision    int       `json:"vision"`
	Score     int       `json:"score"`
	Dead      bool      `json:"dead"`
	InRange   []struct {
		Name      string `json:"name"`
		X         int    `json:"x"`
		Y         int    `json:"y"`
		Direction int    `json:"direction"`
	} `json:"robots_in_range"`

	// On error
	At  string `json:"at"`
	Msg string `json:"msg"`
}

func (r *Robot) String() string {
	if r.At == "error" {
		return fmt.Sprintf("error=%s", r.Msg)
	}
	return fmt.Sprintf("name=%s x=%d y=%d direction=%d score=%d dead=%t in-range=%v", r.Name, r.X, r.Y, r.Direction, r.Score, r.Dead, r.InRange)
}
