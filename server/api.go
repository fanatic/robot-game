package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const password = ""

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

type f func(w http.ResponseWriter, r *http.Request)

func getState(s *State) f {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(s)
	}
}

func postRobot(s *State) f {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := struct {
			Name string `json:"name"`
		}{}
		json.NewDecoder(r.Body).Decode(&payload)

		robot, err := s.NewRobot(payload.Name)
		if err != nil {
			json.NewEncoder(w).Encode(err)
		}
		json.NewEncoder(w).Encode(robot)
	}
}

func deleteRobot(s *State) f {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		err := s.DeleteRobot(id)
		if err != nil {
			json.NewEncoder(w).Encode(err)
		}
		w.WriteHeader(204)
	}
}

func getRobot(s *State) f {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		robot, err := s.Robot(id)
		if err != nil {
			json.NewEncoder(w).Encode(err)
		}
		json.NewEncoder(w).Encode(robot)
	}
}

func postMove(s *State) f {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		if err := s.Move(id); err != nil {
			json.NewEncoder(w).Encode(err)
		}

		robot, err := s.Robot(id)
		if err != nil {
			json.NewEncoder(w).Encode(err)
		}
		json.NewEncoder(w).Encode(robot)
	}
}

func postTurn(s *State) f {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		payload := struct {
			Direction bool `json:"direction"`
		}{}
		json.NewDecoder(r.Body).Decode(&payload)

		if err := s.Turn(id, payload.Direction); err != nil {
			json.NewEncoder(w).Encode(err)
		}

		robot, err := s.Robot(id)
		if err != nil {
			json.NewEncoder(w).Encode(err)
		}
		json.NewEncoder(w).Encode(robot)
	}
}

func postAttack(s *State) f {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		if err := s.Attack(id); err != nil {
			json.NewEncoder(w).Encode(err)
		}

		robot, err := s.Robot(id)
		if err != nil {
			json.NewEncoder(w).Encode(err)
		}
		json.NewEncoder(w).Encode(robot)
	}
}

func New() (http.Handler, error) {
	s, err := NewState()
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../client/build")))

	r.HandleFunc("/state"+password, getState(s)).Methods("GET")
	r.HandleFunc("/robot", postRobot(s)).Methods("POST")
	r.HandleFunc("/robot/{id}", getRobot(s)).Methods("GET")
	r.HandleFunc("/robot/{id}/move", postMove(s)).Methods("POST")
	r.HandleFunc("/robot/{id}/turn", postTurn(s)).Methods("POST")
	r.HandleFunc("/robot/{id}/attack", postAttack(s)).Methods("POST")
	r.HandleFunc("/robot/{id}", deleteRobot(s)).Methods("DELETE")

	return r, nil
}
