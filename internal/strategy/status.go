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

func (ss *StratStates) PPrint() {
	bytes, _ := json.MarshalIndent(ss.states, "", "\t")
	fmt.Println(string(bytes))
}

func (ss *StratStates) Submit(stratname string) {}

func (ss *StratStates) StatusByName(stratname string) *stratState {
	return nil
}

func (ss *StratStates) UpdateByID(id string) {}
