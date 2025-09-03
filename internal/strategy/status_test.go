package strategy

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/jamesonhm/gochain/internal/dt"
)

func TestStatusSubmit(t *testing.T) {
	stratstates := NewStratStates("test_states.json")
	submit_time := time.Date(2025, 8, 1, 13, 0, 0, 0, dt.TZNY())
	stratstates.Submit("test_strat", submit_time)

	assert.Equal(t, stratstates.StatusByName("test_strat").lastSubmitted, submit_time)
}
