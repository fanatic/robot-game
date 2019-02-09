package server

import (
	"github.com/asdine/storm"
)

// Game holds variables used for lifetime of the API process
type Game struct {
	db *storm.DB
}

// NewGame returns a Game struct with db handle
func NewGame(dbPath string) (*Game, error) {
	db, err := storm.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &Game{db: db}, nil
}

// Close a db
func (g *Game) Close() error {
	return g.db.Close()
}
