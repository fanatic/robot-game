package server

import "math/rand"

// Directional names for robot direction
const (
	North = iota
	East
	South
	West
)

// Location is a coordinate on the grid
type Location struct {
	X int
	Y int
}

func adjacentGridLocations(x, y int) map[int]Location {
	return map[int]Location{
		South: Location{x, y + 1},
		East:  Location{x + 1, y},
		North: Location{x, y - 1},
		West:  Location{x - 1, y},
	}
}

func (s *State) randomFreeLocation() (x int, y int, direction int) {
	for i := 0; i < s.Grid*s.Grid; i++ {
		x = rand.Intn(s.Grid)
		y = rand.Intn(s.Grid)
		direction = rand.Intn(4)

		if r := s.locateRobot(x, y); r == nil || r.Dead {
			return
		}
	}
	return
}
