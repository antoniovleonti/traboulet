package kuba

type Direction int
const (
  DirNil Direction = iota
  DirUp
  DirDown
  DirRight
  DirLeft
)

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

type lastMoveT struct {
  x int
  y int
  d Direction
  length int
}

