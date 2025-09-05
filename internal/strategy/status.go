package strategy

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type StratStates struct {
	states map[string]stratState
	mu     sync.RWMutex
	fname  string
}

type stratState struct {
	name          string
	lastSubmitted time.Time
	orderDetails  []orderDetail
}

type orderDetail struct {
	id     string
	status string
}

func NewStratStates(fname string) *StratStates {
	states := make(map[string]stratState)
	if data, err := os.ReadFile(fname); err != nil {
		_, err := os.Create(fname)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if len(data) > 0 {
			if err := json.Unmarshal(data, &states); err != nil {
				log.Fatal(err)
			}
		}
	}
	return &StratStates{
		states: states,
		fname:  fname,
	}

}

func newStratState(ts time.Time) stratState {
	return stratState{
		lastSubmitted: ts,
		orderDetails:  make([]orderDetail, 0),
	}
}

func (ss *StratStates) PPrint() {
	bytes, _ := json.MarshalIndent(ss.states, "", "\t")
	fmt.Println(string(bytes))
}

func (ss *StratStates) Submit(stratname string, ts time.Time) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if state, ok := ss.states[stratname]; !ok {
		ss.states[stratname] = newStratState(ts)
	} else {
		state.lastSubmitted = ts
		ss.states[stratname] = state
	}
	// TODO: write back to file
}

func (ss *StratStates) StatusByName(stratname string) *stratState {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if state, ok := ss.states[stratname]; ok {
		return &state
	}
	// TODO: Complete this
	return nil
}

func (ss *StratStates) UpdateByID(id string) {}
