package kuba

import "encoding/json"

type Direction int
const (
  DirNil Direction = iota
  DirUp
  DirDown
  DirRight
  DirLeft
)

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
  x int
  y int
  d Direction
}

func (m Move) dx() int {
  return m.d.dx()
}

func (m Move) dy() int {
  return m.d.dy()
}


func directionFromDxDy(dx, dy int) (bool, Direction) {
  if dx == 0 && dy == -1 {
    return true, DirUp
  } else if dx == 0 && dy == 1 {
    return true, DirDown
  } else if dx == 1 && dy == 0 {
    return true, DirRight
  } else if dx == -1 && dy == 0 {
    return true, DirLeft
  } else {
    return false, DirNil
  }
}

func (m Move) MarshalJSON() ([]byte, error) {
  type publicMove struct {
    X int `json:"x"`
    Y int `json:"y"`
    D string `json:"d"`
  }

  return json.Marshal(publicMove{
    X: m.x,
    Y: m.y,
    D: m.d.String(),
  })
}

type LastMoveT struct {
  x int
  y int
  d Direction
  length int
}

func (m LastMoveT) MarshalJSON() ([]byte, error) {
  type publicLastMoveT struct {
    X int `json:"x"`
    Y int `json:"y"`
    D string `json:"d"`
    Length int `json:"length"`
  }

  return json.Marshal(publicLastMoveT{
    X: m.x,
    Y: m.y,
    D: m.d.String(),
    Length: m.length,
  })
}
