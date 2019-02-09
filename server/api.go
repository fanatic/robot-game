package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const password = ""

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/state" && r.Method == "GET" {
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

func getState(s *State, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	st, err := s.RefreshState()
	if err != nil {
		return nil, err
	}
	for i := range st.Robots {
		st.Robots[i].ID = ""
	}
	return st, nil
}

func postRobot(s *State, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	payload := struct {
		Name string `json:"name"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid payload body {\"name\": \"JP\"}")
	}

	robot, err := s.NewRobot(payload.Name)
	if err != nil {
		return nil, err
	}

	return robot, nil
}

func deleteRobot(s *State, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	err := s.DeleteRobot(id)
	if err != nil {
		return nil, err
	}

	w.WriteHeader(204)
	return nil, nil
}

func getRobot(s *State, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	robot, err := s.Robot(id)
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}
	return robot, nil
}

func postMove(s *State, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	if err := s.Move(id); err != nil {
		return nil, err
	}

	robot, err := s.Robot(id)
	if err != nil {
		return nil, err
	}
	return robot, nil
}

func postTurn(s *State, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	payload := struct {
		Direction bool `json:"direction"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid payload body {\"direction\": true}")
	}

	if err := s.Turn(id, payload.Direction); err != nil {
		return nil, err
	}

	robot, err := s.Robot(id)
	if err != nil {
		return nil, err
	}
	return robot, nil
}

func postAttack(s *State, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	if err := s.Attack(id); err != nil {
		return nil, err
	}

	robot, err := s.Robot(id)
	if err != nil {
		return nil, err
	}
	return robot, nil
}

type f func(s *State, w http.ResponseWriter, r *http.Request) (interface{}, error)
type hf func(w http.ResponseWriter, r *http.Request)

func handlerWrapper(s *State, f f) hf {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		m, err := f(s, w, r)
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

func New(s *State) (http.Handler, error) {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	r.HandleFunc("/state"+password, handlerWrapper(s, getState)).Methods("GET")
	r.HandleFunc("/robots", handlerWrapper(s, postRobot)).Methods("POST")
	r.HandleFunc("/robots/{id}", handlerWrapper(s, getRobot)).Methods("GET")
	r.HandleFunc("/robots/{id}/move", handlerWrapper(s, postMove)).Methods("POST")
	r.HandleFunc("/robots/{id}/turn", handlerWrapper(s, postTurn)).Methods("POST")
	r.HandleFunc("/robots/{id}/attack", handlerWrapper(s, postAttack)).Methods("POST")
	r.HandleFunc("/robots/{id}", handlerWrapper(s, deleteRobot)).Methods("DELETE")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../client/build")))

	return r, nil
}
