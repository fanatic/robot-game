package server

import (
	"time"

	"github.com/asdine/storm"
)

//const actionDelay = 30 * time.Second
const actionDelay = 30 * time.Millisecond
const robotLimit = 1

// State saves the current round to the db to allow for restarts
type State struct {
	// Saved values
	ID    int `json:"-"`
	Round int `json:"round"`

	// Values not saved
	Grid   int     `json:"grid"`
	Robots []Robot `json:"robots"`

	// Values returned to UI only
	CurrentDelay      time.Duration `json:"delay"`
	CurrentRobotLimit int           `json:"robot_limit"`
}

// State returns state from db
func (g *Game) State() (*State, error) {
	var st []State
	if err := g.db.All(&st); err != nil && err != storm.ErrNotFound {
		return nil, err
	}
	var state State
	if len(st) == 1 {
		state = st[0]
	}

	var robots = make([]Robot, 0)
	if err := g.db.All(&robots); err != nil && err != storm.ErrNotFound {
		return nil, err
	}

	state.Grid = 16
	state.Robots = robots
	state.CurrentDelay = actionDelay
	state.CurrentRobotLimit = robotLimit

	return &state, nil
}

// UpdateRound checks if the round is over, and starts a new one
func (g *Game) UpdateRound() error {
	s, err := g.State()
	if err != nil {
		return err
	}

	if len(s.Robots) <= 1 {
		return nil
	}

	alive := s.robotsAlive()
	if len(alive) != 1 {
		return nil
	}

	// Round Over!

	s.ID = 1 // hardcode id so there can only be one state
	s.Round = s.Round + 1
	if err := g.db.Save(s); err != nil {
		return err
	}

	for _, robot := range s.Robots {
		robot.Dead = false // see next comment
		robot.X, robot.Y, robot.Direction = s.randomFreeLocation()
		if robot.ID == alive[0].ID {
			// Winner Winner, Chicken Dinner
			robot.Score += 100
		}
		if err := g.db.Update(&robot); err != nil {
			return err
		}
		// Update won't save zero-value fields, so do it explicitly for dead and ignore issues with x/y/direction being 0
		if err := g.db.UpdateField(&robot, "Dead", false); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) robotsAlive() []Robot {
	alive := []Robot{}
	for _, robot := range s.Robots {
		if !robot.Dead {
			alive = append(alive, robot)
		}
	}
	return alive
}

func (s *State) locateRobot(x, y int) *Robot {
	for _, robot := range s.Robots {
		if robot.X == x && robot.Y == y {
			return &robot
		}
	}

	return nil
}

// RobotsInRange returns a list of robots within the vision of the current robot
func (s *State) RobotsInRange(r *Robot) []ShortRobot {
	// TODO(jp): support more than vision=4
	inRange := []ShortRobot{}
	for _, l := range adjacentGridLocations(r.X, r.Y) {
		if robot := s.locateRobot(l.X, l.Y); robot != nil && !robot.Dead {
			inRange = append(inRange, ShortRobot{Name: robot.Name, X: robot.X, Y: robot.Y, Direction: robot.Direction})
		}
	}
	return inRange
}
