package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func postRobot(g *Game, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	payload := struct {
		Name string `json:"name"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid payload body {\"name\": \"JP\"}")
	}
	if len(payload.Name) != 2 {
		return nil, fmt.Errorf("name must be exactly 2 characters")
	}

	robot, err := g.NewRobot(strings.ToUpper(payload.Name))
	if err != nil {
		return nil, err
	}

	return robot, nil
}

func deleteRobot(g *Game, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	err := g.DeleteRobot(id)
	if err != nil {
		return nil, err
	}

	w.WriteHeader(204)
	return nil, nil
}

func getRobot(g *Game, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	robot, err := g.Robot(id)
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}
	return robot, nil
}

func postMove(g *Game, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	if err := g.Move(id); err != nil {
		return nil, err
	}

	robot, err := g.Robot(id)
	if err != nil {
		return nil, err
	}
	return robot, nil
}

func postTurn(g *Game, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]

	payload := struct {
		Direction bool `json:"direction"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid payload body {\"direction\": true}")
	}

	if err := g.Turn(id, payload.Direction); err != nil {
		return nil, err
	}

	robot, err := g.Robot(id)
	if err != nil {
		return nil, err
	}
	return robot, nil
}

func postAttack(g *Game, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := mux.Vars(r)["id"]
	if err := g.Attack(id); err != nil {
		return nil, err
	}

	robot, err := g.Robot(id)
	if err != nil {
		return nil, err
	}
	return robot, nil
}
