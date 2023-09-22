package kuba

import (
	"testing"
	"time"
)

func TestAgentToMarble(t *testing.T) {
	if agentWhite.marble() != marbleWhite {
		t.Fail()
	}
	if agentBlack.marble() != marbleBlack {
		t.Fail()
	}
	if agentNil.marble() != marbleNil {
		t.Fail()
	}
}

func TestOtherAgent(t *testing.T) {
	if agentWhite.otherAgent() != agentBlack {
		t.Fail()
	}
	if agentBlack.otherAgent() != agentWhite {
		t.Fail()
	}
}

func TestAgentToWinStatus(t *testing.T) {
	if agentWhite.winStatus() != statusWhiteWon {
		t.Fail()
	}
	if agentBlack.winStatus() != statusBlackWon {
		t.Fail()
	}
}

func TestStartTurnCallback(t *testing.T) {
	a := agent{
		time: time.Nanosecond,
	}

	done := make(chan bool)
	cb := func() {
		close(done)
	}

	a.startTurn(cb)
	<-done
}

func TestEndTurn(t *testing.T) {
	totalTime := time.Millisecond * 100
	a := agent{
		time: totalTime,
	}

	cbDone := make(chan bool)
	cb := func() {
		close(cbDone)
	}

	timerDone := time.NewTimer(totalTime + time.Millisecond*100)

	a.startTurn(cb)
	a.endTurn()

	select {
	case <-cbDone:
		t.Error("The callback fired even though we ended the turn")
	case <-timerDone.C:
	}
}
