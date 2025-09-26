package strategy

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/jamesonhm/gochain/internal/dt"
	"github.com/jamesonhm/gochain/internal/tasty"
)

func TestStatusSubmit(t *testing.T) {
	stratstates := NewStatus("test_states.json")
	submit_time := time.Date(2025, 8, 1, 13, 0, 0, 0, dt.TZNY())
	order := tasty.NewOrder{}
	stratstates.Submit("test_strat", submit_time, "1", order)

	status, err := stratstates.StatusByName("test_strat")
	assert.Equal(t, err, nil)
	assert.Equal(t, status.LastSubmitted, submit_time)
}
