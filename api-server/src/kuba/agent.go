package kuba

import (
	"time"
)

type AgentColor int

const (
	agentNil AgentColor = iota
	agentWhite
	agentBlack
)

type agent struct {
	score    int
	time     time.Duration
	deadline *time.Time
	timer    *time.Timer
}

func (ac AgentColor) String() string {
	if ac == agentWhite {
		return "WHITE"
	} else if ac == agentBlack {
		return "BLACK"
	} else {
		panic("invalid AgentColor!")
	}
}

func (ac AgentColor) marble() Marble {
	return Marble(ac)
}

func (ac AgentColor) otherAgent() AgentColor {
	return ac%2 + 1
}

func (ac AgentColor) winStatus() Status {
	return Status(ac)
}

func (a *agent) startTurn(timeoutCb func()) bool {
	if a == nil {
		return true
	}
	if a.timer != nil || a.deadline != nil {
		return false
	}
	tmp := time.Now().Add(a.time)
	a.deadline = &tmp
	a.timer = time.AfterFunc(a.time, timeoutCb)
	return true
}

func (a *agent) endTurn() bool {
	if a == nil {
		return true
	}
	if a.timer != nil && !a.timer.Stop() {
		return false
	}
	a.timer = nil
	if a.deadline != nil {
		a.time = time.Until(*a.deadline)
		a.deadline = nil
	}
	return true
}
