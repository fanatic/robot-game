package server

import "net/http"

func getState(g *Game, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	s, err := g.State()
	if err != nil {
		return nil, err
	}
	for i := range s.Robots {
		s.Robots[i].ID = ""
	}
	return s, nil
}
