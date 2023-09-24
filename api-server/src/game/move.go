package game

import (
	"encoding/json"
	"errors"
)

type Direction int

const (
	DirNil Direction = iota
	DirUp
	DirDown
	DirRight
	DirLeft
)

func DirectionFromString(s string) (Direction, error) {
  for _, d := range []Direction{DirUp, DirDown, DirRight, DirLeft} {
    if s == d.String() {
      return d, nil
    }
  }
  return DirNil, errors.New("invalid direction!")
}

func (d *Direction) UnmarshalJSON(raw []byte) error {
	if string(raw) == "" {
		return errors.New("Direction cannot be empty.")
	}
	tmp, err := DirectionFromString(string(raw))
	*d = tmp
	return err
}

func (d Direction) String() string {
	if d == DirUp {
		return "UP"
	} else if d == DirDown {
		return "DOWN"
	} else if d == DirRight {
		return "RIGHT"
	} else if d == DirLeft {
		return "LEFT"
	} else {
		panic("invalid direction!")
	}
}

func (d Direction) dx() int {
	if d == DirUp || d == DirDown {
		return 0
	} else if d == DirRight {
		return 1
	} else if d == DirLeft {
		return -1
	} else {
		panic("invalid direction!")
	}
}

func (d Direction) dy() int {
	if d == DirLeft || d == DirRight {
		return 0
	} else if d == DirUp {
		return -1
	} else if d == DirDown {
		return 1
	} else {
		panic("invalid direction!")
	}
}

func (d Direction) isValid() bool {
	return d > DirNil && d <= DirLeft
}

func (d Direction) reverse() Direction {
	if d == DirUp {
		return DirDown
	} else if d == DirDown {
		return DirUp
	} else if d == DirRight {
		return DirLeft
	} else if d == DirLeft {
		return DirRight
	} else {
		panic("invalid direction!")
	}
}

type Move struct {
	X int       `json:"x"`
	Y int       `json:"y"`
	D Direction `json:"d,string"`
}

func (m Move) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		X int    `json:"x"`
		Y int    `json:"y"`
		D string `json:"d"`
	}{
		X: m.X,
		Y: m.Y,
		D: m.D.String(),
	})
}

func (m Move) dx() int {
	return m.D.dx()
}

func (m Move) dy() int {
	return m.D.dy()
}

type MoveWMarblesMoved struct {
	X            int       `json:"x"`
	Y            int       `json:"y"`
	D            Direction `json:"d,string"`
	MarblesMoved int       `json:"marblesMoved"`
}

func (m MoveWMarblesMoved) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		X            int    `json:"x"`
		Y            int    `json:"y"`
		D            string `json:"d"`
		MarblesMoved int    `json:"marblesMoved"`
	}{
		X:            m.X,
		Y:            m.Y,
		D:            m.D.String(),
		MarblesMoved: m.MarblesMoved,
	})
}
