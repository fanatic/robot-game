package server

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Robot struct {
	ID        string       `json:"id,omitempty"` // also the secret
	CreatedAt time.Time    `json:"created_at"`
	Name      string       `json:"name"`
	X         int          `json:"x"`
	Y         int          `json:"y"`
	Color     string       `json:"color"`
	Direction int          `json:"direction"`
	Vision    int          `json:"vision"`
	Score     int          `json:"score"`
	Dead      bool         `json:"dead"`
	InRange   []ShortRobot `json:"robots_in_range"`
}

type ShortRobot struct {
	Name      string `json:"name"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction int    `json:"direction"`
}

func (s *State) NewRobot(name string) (*Robot, error) {
	colors := []string{"#e6194b", "#3cb44b", "#ffe119", "#4363d8", "#f58231", "#911eb4", "#46f0f0", "#f032e6", "#bcf60c", "#fabebe", "#008080", "#e6beff", "#9a6324", "#fffac8", "#800000", "#aaffc3", "#808000", "#ffd8b1", "#000075", "#808080", "#ffffff", "#000000"}

	newState, err := s.RefreshState()
	if err != nil {
		return nil, err
	}

	// Limit robots by name
	robotCount := 0
	for _, robot := range newState.Robots {
		if robot.Name == name {
			robotCount++
		}
	}
	if robotCount >= newState.CurrentRobotLimit {
		return nil, fmt.Errorf("No more robots - you're at the limit")
	}

	// Find first unused color
	var unusedColor string
	for _, color := range colors {
		isUsed := false
		for _, robot := range newState.Robots {
			if robot.Color == color {
				isUsed = true
			}
		}
		if !isUsed {
			unusedColor = color
			break
		}
	}

	x := rand.Intn(s.Grid)
	y := rand.Intn(s.Grid)
	direction := rand.Intn(4)
	id, _ := uuid.NewRandom()
	r := Robot{
		ID:        id.String(),
		CreatedAt: time.Now(),
		X:         x,
		Y:         y,
		Color:     unusedColor,
		Name:      name,
		Direction: direction,
		Vision:    4,
		Score:     0,
	}

	if err := s.db.Save(&r); err != nil {
		return nil, err
	}

	newState, err = s.RefreshState()
	if err != nil {
		return nil, err
	}
	*s = *newState

	r.InRange = s.RobotsInRange(&r)

	return &r, nil
}

const (
	North = iota
	East
	South
	West
)

func (s *State) Robot(id string) (*Robot, error) {
	var r Robot
	if err := s.db.One("ID", id, &r); err != nil {
		return nil, err
	}

	if r.Dead {
		return nil, fmt.Errorf("this robot be dead")
	}

	newState, err := s.RefreshState()
	if err != nil {
		return nil, err
	}
	r.InRange = newState.RobotsInRange(&r)

	return &r, nil
}

func (s *State) DeleteRobot(id string) error {
	return s.db.DeleteStruct(&Robot{ID: id})
}

func (s *State) Move(id string) error {
	time.Sleep(s.CurrentDelay)

	r, err := s.Robot(id)
	if err != nil {
		return err
	}

	newState, err := s.RefreshState()
	if err != nil {
		return err
	}

	switch r.Direction {
	case North:
		if r.Y-1 < 0 {
			return fmt.Errorf("off the grid")
		}
		if robot := newState.locateRobot(r.X, r.Y-1); robot != nil && !robot.Dead {
			return fmt.Errorf("something's in the way")
		}
		return s.db.UpdateField(r, "Y", r.Y-1)
	case East:
		if r.X+1 == s.Grid {
			return fmt.Errorf("off the grid")
		}
		if robot := newState.locateRobot(r.X+1, r.Y); robot != nil && !robot.Dead {
			return fmt.Errorf("something's in the way")
		}
		return s.db.UpdateField(r, "X", r.X+1)
	case South:
		if r.Y+1 == s.Grid {
			return fmt.Errorf("off the grid")
		}
		if robot := newState.locateRobot(r.X, r.Y+1); robot != nil && !robot.Dead {
			return fmt.Errorf("something's in the way")
		}
		return s.db.UpdateField(r, "Y", r.Y+1)
	case West:
		if r.X-1 < 0 {
			return fmt.Errorf("off the grid")
		}
		if robot := newState.locateRobot(r.X-1, r.Y); robot != nil && !robot.Dead {
			return fmt.Errorf("something's in the way")
		}
		return s.db.UpdateField(r, "X", r.X-1)
	}
	return fmt.Errorf("unknown direction")
}

func (s *State) Turn(id string, direction bool) error {
	time.Sleep(s.CurrentDelay)

	r, err := s.Robot(id)
	if err != nil {
		return err
	}

	switch {
	case (r.Direction == North && direction) || (r.Direction == South && !direction):
		return s.db.UpdateField(r, "Direction", West)
	case (r.Direction == East && direction) || (r.Direction == West && !direction):
		return s.db.UpdateField(r, "Direction", North)
	case (r.Direction == South && direction) || (r.Direction == North && !direction):
		return s.db.UpdateField(r, "Direction", East)
	case (r.Direction == West && direction) || (r.Direction == East && !direction):
		return s.db.UpdateField(r, "Direction", South)
	}
	return fmt.Errorf("unknown direction")
}

func (s *State) Attack(id string) error {
	time.Sleep(s.CurrentDelay)

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
		if err := s.db.UpdateField(r, "Score", r.Score+10); err != nil {
			return err
		}
		if err := s.db.UpdateField(robot, "Dead", true); err != nil {
			return err
		}
		return s.UpdateRound()
	}

	return fmt.Errorf("swwwing and a missss")
}

func (s *State) locateRobot(x, y int) *Robot {
	for _, robot := range s.Robots {
		if robot.X == x && robot.Y == y {
			return &robot
		}
	}

	return nil
}

func (s *State) RobotsInRange(r *Robot) []ShortRobot {
	// TODO(jp): support more than vision=4
	inRange := []ShortRobot{}
	if robot := s.locateRobot(r.X, r.Y+1); robot != nil && !robot.Dead {
		inRange = append(inRange, ShortRobot{Name: robot.Name, X: robot.X, Y: robot.Y, Direction: robot.Direction})
	} else if robot := s.locateRobot(r.X, r.Y-1); robot != nil && !robot.Dead {
		inRange = append(inRange, ShortRobot{Name: robot.Name, X: robot.X, Y: robot.Y, Direction: robot.Direction})
	} else if robot := s.locateRobot(r.X+1, r.Y); robot != nil && !robot.Dead {
		inRange = append(inRange, ShortRobot{Name: robot.Name, X: robot.X, Y: robot.Y, Direction: robot.Direction})
	} else if robot := s.locateRobot(r.X-1, r.Y); robot != nil && !robot.Dead {
		inRange = append(inRange, ShortRobot{Name: robot.Name, X: robot.X, Y: robot.Y, Direction: robot.Direction})
	}
	return inRange
}
