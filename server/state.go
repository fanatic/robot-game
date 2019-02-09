package server

import (
	"math/rand"
	"time"

	"github.com/asdine/storm"
)

const actionDelay = 30 * time.Second
const robotLimit = 1

type State struct {
	// Saved values
	ID    int `json:"-"`
	Round int `json:"round"`

	// Values not saved
	Grid              int           `json:"grid"`
	CurrentDelay      time.Duration `json:"delay"`
	CurrentRobotLimit int           `json:"robot_limit"`
	Robots            []Robot       `json:"robots"`
	db                *storm.DB
}

func NewState(dbPath string) (*State, error) {
	db, err := storm.Open(dbPath)
	if err != nil {
		return nil, err
	}
	s := &State{db: db}
	return s.RefreshState()
}

func (s *State) Close() error {
	return s.db.Close()
}

func (s *State) RefreshState() (*State, error) {
	var st []State
	if err := s.db.All(&st); err != nil && err != storm.ErrNotFound {
		return nil, err
	}
	var state State
	if len(st) == 1 {
		state = st[0]
	}

	var robots = make([]Robot, 0, 0)
	if err := s.db.All(&robots); err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	state.Grid = 16
	state.Robots = robots
	state.db = s.db
	state.CurrentDelay = actionDelay
	state.CurrentRobotLimit = robotLimit

	return &state, nil
}

func (s *State) UpdateRound() error {
	newState, err := s.RefreshState()
	if err != nil {
		return err
	}

	if len(newState.Robots) <= 1 {
		return nil
	}

	// Slice of robots still alive
	alive := []Robot{}
	for _, robot := range newState.Robots {
		if !robot.Dead {
			alive = append(alive, robot)
		}
	}
	if len(alive) != 1 {
		return nil
	}

	// Round Over!

	newState.ID = 1 // hardcode id so there can only be one state
	newState.Round = newState.Round + 1
	if err := s.db.Save(newState); err != nil {
		return err
	}

	for _, robot := range newState.Robots {
		robot.Dead = false // see next comment
		robot.X = rand.Intn(newState.Grid)
		robot.Y = rand.Intn(newState.Grid)
		robot.Direction = rand.Intn(4)
		if robot.ID == alive[0].ID {
			// Winner Winner, Chicken Dinner
			robot.Score += 100
		}
		if err := s.db.Update(&robot); err != nil {
			return err
		}
		// Update won't save zero-value fields, so do it explicitly for dead and ignore issues with x/y/direction being 0
		if err := s.db.UpdateField(&robot, "Dead", false); err != nil {
			return err
		}
	}

	newState, err = s.RefreshState()
	if err != nil {
		return err
	}
	*s = *newState

	return nil
}
