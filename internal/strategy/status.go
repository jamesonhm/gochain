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
	Name          string
	LastSubmitted time.Time
	OrderDetails  []orderDetail
}

type orderDetail struct {
	Id     string
	Status string
}

func NewStratStates(filename string) *StratStates {
	states := make(map[string]stratState)
	if data, err := os.ReadFile(filename); err != nil {
		_, err := os.Create(filename)
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
		fname:  filename,
	}

}

func newStratState(ts time.Time) stratState {
	return stratState{
		LastSubmitted: ts,
		OrderDetails:  make([]orderDetail, 0),
	}
}

func (ss *StratStates) PPrint() {
	bytes, _ := json.MarshalIndent(ss.states, "", "\t")
	fmt.Println("Strategy States:")
	fmt.Println(string(bytes))
}

func (ss *StratStates) Submit(stratname string, ts time.Time) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if state, ok := ss.states[stratname]; !ok {
		ss.states[stratname] = newStratState(ts)
	} else {
		state.LastSubmitted = ts
		ss.states[stratname] = state
	}

	byt, err := json.MarshalIndent(ss.states, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(ss.fname, byt, 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func (ss *StratStates) StatusByName(stratname string) (stratState, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if state, ok := ss.states[stratname]; ok {
		return state, nil
	}
	return stratState{}, fmt.Errorf("No status for strategy name")
}

func (ss *StratStates) LastSubmitted(stratname string) (time.Time, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if state, ok := ss.states[stratname]; ok {
		return state.LastSubmitted, nil
	}
	return time.Now().AddDate(-1, 0, 0), fmt.Errorf("No status for strategy name")
}
func (ss *StratStates) UpdateByID(id string) {}
