package kuba

import (
  "encoding/json"
  "testing"
)

func TestUnmarshalMove(t *testing.T) {
  type testcase struct {
    raw []byte
    expected Move
  }
  testcases := []testcase{
    { raw: []byte(`{ "X": 0, "Y": 0, "D": "UP" }`),
      expected: Move{ X: 0, Y: 0, D: DirUp } },
    { raw: []byte(`{ "X": 7, "Y": 7, "D": "DOWN" }`),
      expected: Move{ X: 7, Y: 7, D: DirDown } },
    { raw: []byte(`{ "X": 0, "Y": 0, "D": "LEFT" }`),
      expected: Move{ X: 0, Y: 0, D: DirLeft } },
    { raw: []byte(`{ "X": 0, "Y": 0, "D": "RIGHT" }`),
      expected: Move{ X: 0, Y: 0, D: DirRight } },
  }

  for idx, tc := range testcases {
    var actual Move
    err := json.Unmarshal(tc.raw, &actual)
    if err != nil {
      t.Errorf("test case %d: %s", idx, err)
    }
    if actual != tc.expected {
      t.Errorf("test case %d: got unexpected result:\n" +
               "actual: %v\nexpected: %v\n", idx, actual, tc.expected)
    }
  }
}

func TestUnmarshalInvalidMove(t *testing.T) {
  testcases := [][]byte{
    []byte(`{ "X": 0, "Y": 0, "D": "BLAH BLAH" }`),
    []byte(`{ "X": 7, "Y": 7, "D": "DOWN " }`),
    []byte(`{ "X": 0, "Y": 0, "D": "left" }`),
    []byte(`{ "X": 0, "Y": 0, "D": "" }`),
  }

  for idx, tc := range testcases {
    var actual Move
    err := json.Unmarshal(tc, &actual)
    if err == nil {
      t.Errorf("test case %d: expected error", idx)
    }
  }
}
