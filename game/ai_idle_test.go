package game

import "testing"

func TestIdleProvokedLeadsToKill(t *testing.T) {
	tt := NewDefaultTransitionTable()
	if tt.GetNextState(StateIdle, EventProvoked) != StateKill {
		t.Fatal("idle + provoked must lead to kill")
	}
}
