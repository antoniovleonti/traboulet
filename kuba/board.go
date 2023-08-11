package kuba

import (
  "encoding/json"
  "errors"
)

type Marble int
const (
  marbleNil Marble = iota
  marbleWhite
  marbleBlack
  marbleRed
)

func (m Marble) String() string {
  if m == marbleNil {
    return " "
  } else if m == marbleWhite {
    return "W"
  } else if m == marbleBlack {
    return "B"
  } else if m == marbleRed {
    return "R"
  } else {
    panic("invalid marble!")
  }
}

func marbleFromString(s string) (Marble, bool) {
  if s == marbleNil.String() {
    return marbleNil, true
  } else if s == marbleWhite.String() {
    return marbleWhite, true
  } else if s == marbleBlack.String() {
    return marbleBlack, true
  } else if s == marbleRed.String() {
    return marbleRed, true
  } else {
    return marbleNil, false
  }
}

type BoardT [][]Marble

func (b BoardT) MarshalJSON() ([]byte, error) {
  s := make([][]string, len(b))
  for i := 0; i < len(b); i++ {
    s[i] = make([]string, len(b[0]))
  }

  for i, row := range b {
    for j, el := range row {
      s[i][j] = el.String()
    }
  }

  return json.Marshal(s)
}

func (b *BoardT) UnmarshalJSON(raw []byte) error {
  if string(raw) == "" {
    return nil
  }
  var strSlice [][]string
  json.Unmarshal(raw, &strSlice)
  *b = make(BoardT, len(strSlice))
  for i := 0; i < len(strSlice); i++ {
    (*b)[i] = make([]Marble, len(strSlice[i]))
    for j := 0; j < len(strSlice[i]); j++ {
      marble, ok := marbleFromString(strSlice[i][j])
      if !ok {
        return errors.New("invalid marble string " + strSlice[i][j])
      }
      (*b)[i][j] = marble
    }
  }
  return nil
}
