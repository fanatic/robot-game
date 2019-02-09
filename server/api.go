package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const password = ""

// New returns a new http Handler for the robot game API
func New(g *Game) (http.Handler, error) {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	routes := map[string]map[string]f{
		"GET": {
			"/state" + password: getState,
			"/robots/{id}":      getRobot,
		},
		"POST": {
			"/robots":             postRobot,
			"/robots/{id}/move":   postMove,
			"/robots/{id}/turn":   postTurn,
			"/robots/{id}/attack": postAttack,
		},
		"DELETE": {
			"/robots/{id}": deleteRobot,
		},
	}

	for method, paths := range routes {
		for path, f := range paths {
			r.Methods(method).Path(path).HandlerFunc(handlerWrapper(g, f))
		}
	}

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../client/build")))

	return r, nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/state"+password && r.Method == "GET" {
			// Skip logging
			next.ServeHTTP(w, r)
			return
		}
		log.Printf("http at=start  method=%s path=%s\n", r.Method, r.RequestURI)
		startTime := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("http at=finish method=%s path=%s duration=%s\n", r.Method, r.RequestURI, time.Since(startTime))
	})
}

type f func(g *Game, w http.ResponseWriter, r *http.Request) (interface{}, error)

func handlerWrapper(g *Game, f f) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		m, err := f(g, w, r)
		if err != nil {
			log.Println(err)
			json.NewEncoder(w).Encode(map[string]string{"at": "error", "msg": err.Error()})
			return
		}
		if m != nil {
			if err := json.NewEncoder(w).Encode(m); err != nil {
				log.Println(err)
				json.NewEncoder(w).Encode(map[string]string{"at": "error", "msg": err.Error()})
				return
			}
		}
	}
}
