package strategy

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
)

type StratStates struct {
	mu     sync.RWMutex
	fname  string
	states states
}

type states struct {
	SeqPfid    int                    `json:"seq-pfid"`
	Strategies map[string]stratOrders `json:"strategies"`
}

type stratOrders struct {
	LastSubmitted time.Time              `json:"last-submitted"`
	OrderDetails  map[string]orderDetail `json:"order-details"`
}

type orderDetail struct {
	OrderId string `json:"order-id"`
	Status  string `json:"status"`
}

func NewStratStates(filename string) *StratStates {
	states := states{
		Strategies: make(map[string]stratOrders),
	}
	if data, err := os.ReadFile(filename); err != nil {
		slog.Info("Unable to read strat states file, creating new")
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

func newStratOrders(ts time.Time, pfid string) stratOrders {
	return stratOrders{
		LastSubmitted: ts,
		OrderDetails: map[string]orderDetail{
			pfid: orderDetail{},
		},
	}
}

func (ss *StratStates) PPrint() {
	bytes, _ := json.MarshalIndent(ss.states, "", "  ")
	fmt.Println("Strategy States:")
	fmt.Println(string(bytes))
}

func (ss *StratStates) Submit(stratname string, ts time.Time, pfid string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if orders, ok := ss.states.Strategies[stratname]; !ok {
		ss.states.Strategies[stratname] = newStratOrders(ts, pfid)
	} else {
		orders.LastSubmitted = ts
		orders.OrderDetails[pfid] = orderDetail{}
		ss.states.Strategies[stratname] = orders
	}

	ss.writefile()
}

func (ss *StratStates) StatusByName(stratname string) (stratOrders, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if orders, ok := ss.states.Strategies[stratname]; ok {
		return orders, nil
	}
	return stratOrders{}, fmt.Errorf("No status for strategy name")
}

func (ss *StratStates) LastSubmitted(stratname string) (time.Time, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if state, ok := ss.states.Strategies[stratname]; ok {
		return state.LastSubmitted, nil
	}
	return time.Now().AddDate(-1, 0, 0), fmt.Errorf("No status for strategy name")
}

func (ss *StratStates) NextPFID() int {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.states.SeqPfid += 1
	ss.writefile()
	return ss.states.SeqPfid
}

func (ss *StratStates) UpdateByID(id string) {}

func (ss *StratStates) writefile() {
	byt, err := json.MarshalIndent(ss.states, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(ss.fname, byt, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
