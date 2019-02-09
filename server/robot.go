package server

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Robot is the player
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

// ShortRobot is used when sharing enemy robots
type ShortRobot struct {
	Name      string `json:"name"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction int    `json:"direction"`
}

// NewRobot creates a new robot, saves to the db, and returns it
func (g *Game) NewRobot(name string) (*Robot, error) {
	s, err := g.State()
	if err != nil {
		return nil, err
	}

	// Limit robots by name
	robotCount := 0
	for _, robot := range s.Robots {
		if robot.Name == name {
			robotCount++
		}
	}
	if robotCount >= robotLimit {
		return nil, fmt.Errorf("no more robots - you're at the limit")
	}

	id, _ := uuid.NewRandom()
	r := Robot{
		ID:        id.String(),
		CreatedAt: time.Now(),
		Color:     findFirstUnusedColor(s.Robots),
		Name:      name,
		Vision:    4,
		Score:     0,
	}
	r.X, r.Y, r.Direction = s.randomFreeLocation()

	if err := g.db.Save(&r); err != nil {
		return nil, err
	}

	s, err = g.State()
	if err != nil {
		return nil, err
	}

	r.InRange = s.RobotsInRange(&r)

	return &r, nil
}

// Robot gets an existing robot directly from the db and returns it
func (g *Game) Robot(id string) (*Robot, error) {
	var r Robot
	if err := g.db.One("ID", id, &r); err != nil {
		return nil, err
	}

	if r.Dead {
		return nil, fmt.Errorf("this robot be dead")
	}

	s, err := g.State()
	if err != nil {
		return nil, err
	}
	r.InRange = s.RobotsInRange(&r)

	return &r, nil
}

// DeleteRobot from the db
func (g *Game) DeleteRobot(id string) error {
	return g.db.DeleteStruct(&Robot{ID: id})
}

// Move a robot in the db
func (g *Game) Move(id string) error {
	time.Sleep(actionDelay)

	r, err := g.Robot(id)
	if err != nil {
		return err
	}

	s, err := g.State()
	if err != nil {
		return err
	}

	l, exists := adjacentGridLocations(r.X, r.Y)[r.Direction]
	if !exists {
		return fmt.Errorf("unknown direction")
	}

	if l.X < 0 || l.X == s.Grid || l.Y < 0 || l.Y == s.Grid {
		return fmt.Errorf("off the grid")
	}

	if robot := s.locateRobot(l.X, l.Y); robot != nil && !robot.Dead {
		return fmt.Errorf("something's in the way")
	}

	if l.X != r.X {
		return g.db.UpdateField(r, "X", l.X)
	} else if l.Y != r.Y {
		return g.db.UpdateField(r, "Y", l.Y)
	}

	return nil
}

// Turn a robot in the db
func (g *Game) Turn(id string, direction bool) error {
	time.Sleep(actionDelay)

	r, err := g.Robot(id)
	if err != nil {
		return err
	}

	newDirection := (r.Direction + 1) % 4
	if direction {
		newDirection = (r.Direction + 3) % 4 // -1
	}
	return g.db.UpdateField(r, "Direction", newDirection)
}

// Attack a robot via the db
func (g *Game) Attack(id string) error {
	time.Sleep(actionDelay)

	r, err := g.Robot(id)
	if err != nil {
		return err
	}

	s, err := g.State()
	if err != nil {
		return err
	}

	l := adjacentGridLocations(r.X, r.Y)
	robot := s.locateRobot(l[r.Direction].X, l[r.Direction].Y)

	if robot != nil {
		if robot.Dead {
			return fmt.Errorf("how rude to attack a dead robot")
		}
		if err := g.db.UpdateField(r, "Score", r.Score+10); err != nil {
			return err
		}
		if err := g.db.UpdateField(robot, "Dead", true); err != nil {
			return err
		}
		return g.UpdateRound()
	}

	return fmt.Errorf("swwwing and a missss")
}

func findFirstUnusedColor(robots []Robot) string {
	colors := []string{"#e6194b", "#3cb44b", "#ffe119", "#4363d8", "#f58231", "#911eb4", "#46f0f0", "#f032e6", "#bcf60c", "#fabebe", "#008080", "#e6beff", "#9a6324", "#fffac8", "#800000", "#aaffc3", "#808000", "#ffd8b1", "#000075", "#808080", "#ffffff", "#000000"}

	for _, color := range colors {
		isUsed := false
		for _, robot := range robots {
			if robot.Color == color {
				isUsed = true
			}
		}
		if !isUsed {
			return color
		}
	}
	return colors[0]
}
