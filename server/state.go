package server

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/asdine/storm"
	"github.com/google/uuid"
)

type State struct {
	Round  int     `json:"round"`
	Grid   int     `json:"grid"`
	Robots []Robot `json:"robots"`
	db     *storm.DB
}

func NewState() (*State, error) {
	db, err := storm.Open("my.db")
	if err != nil {
		return nil, err
	}
	s := &State{db: db}
	return s.RefreshState()
}

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

func (s *State) NewRobot(name string) (*Robot, error) {
	colors := []string{"#e6194b", "#3cb44b", "#ffe119", "#4363d8", "#f58231", "#911eb4", "#46f0f0", "#f032e6", "#bcf60c", "#fabebe", "#008080", "#e6beff", "#9a6324", "#fffac8", "#800000", "#aaffc3", "#808000", "#ffd8b1", "#000075", "#808080", "#ffffff", "#000000"}

	x := rand.Intn(s.Grid)
	y := rand.Intn(s.Grid)
	direction := rand.Intn(4)
	id, _ := uuid.NewRandom()
	r := Robot{
		ID:        id.String(),
		X:         x,
		Y:         y,
		Color:     colors[len(s.Robots)%len(colors)],
		Name:      name,
		Direction: direction,
		Vision:    4,
		Score:     0,
	}

	err := s.db.From("robots").Save(r)
	return &r, err
}

const (
	North = iota
	East
	South
	West
)

func (s *State) RefreshState() (*State, error) {
	var state State
	if err := s.db.From("state").Select().First(&state); err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	var robots []Robot
	if err := s.db.From("robots").Select().Find(&robots); err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	state.Grid = 16
	state.Robots = robots

	return &state, nil
}

func (s *State) Robot(id string) (*Robot, error) {
	var r Robot
	if err := s.db.From("robots").One("id", id, &r); err != nil {
		return nil, err
	}

	if r.Dead {
		return nil, fmt.Errorf("this robot be dead")
	}

	return &r, nil
}

func (s *State) DeleteRobot(id string) error {
	return s.db.From("robots").DeleteStruct(Robot{ID: id})
}

func (s *State) Move(id string) error {
	r, err := s.Robot(id)
	if err != nil {
		return err
	}

	switch r.Direction {
	case North:
		if r.Y-1 < 0 {
			return fmt.Errorf("off the grid")
		}
		return s.db.UpdateField(r, "y", r.Y-1)
	case East:
		if r.X+1 == s.Grid {
			return fmt.Errorf("off the grid")
		}
		return s.db.UpdateField(r, "x", r.X+1)
	case South:
		if r.Y+1 == s.Grid {
			return fmt.Errorf("off the grid")
		}
		return s.db.UpdateField(r, "y", r.Y+1)
	case West:
		if r.X-1 < 0 {
			return fmt.Errorf("off the grid")
		}
		return s.db.UpdateField(r, "x", r.X-1)
	}
	return fmt.Errorf("unknown direction")
}

func (s *State) Turn(id string, direction bool) error {
	r, err := s.Robot(id)
	if err != nil {
		return err
	}

	switch {
	case (r.Direction == North && direction) || (r.Direction == South && !direction):
		return s.db.UpdateField(r, "direction", West)
	case (r.Direction == East && direction) || (r.Direction == West && !direction):
		return s.db.UpdateField(r, "direction", North)
	case (r.Direction == South && direction) || (r.Direction == North && !direction):
		return s.db.UpdateField(r, "direction", East)
	case (r.Direction == West && direction) || (r.Direction == East && !direction):
		return s.db.UpdateField(r, "direction", South)
	}
	return fmt.Errorf("unknown direction")
}

func (s *State) Attack(id string) error {
	r, err := s.Robot(id)
	if err != nil {
		return err
	}

	newState, err := s.RefreshState()
	if err != nil {
		return err
	}

	var robot *Robot
	switch r.Direction {
	case North:
		robot = newState.locateRobot(r.X, r.Y-1)
	case East:
		robot = newState.locateRobot(r.X+1, r.Y)
	case South:
		robot = newState.locateRobot(r.X, r.Y+1)
	case West:
		robot = newState.locateRobot(r.X-1, r.Y)
	}

	if robot != nil {
		if robot.Dead {
			return fmt.Errorf("how rude to attack a dead robot")
		}
		if err := s.db.UpdateField(r, "score", r.Score+10); err != nil {
			return err
		}
		return s.db.UpdateField(robot, "dead", true)
	}

	return nil
}

func (s *State) locateRobot(x, y int) *Robot {
	for _, robot := range s.Robots {
		if robot.X == x && robot.Y == y {
			return &robot
		}
	}

	return nil
}
