package strategy

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/jamesonhm/gochain/internal/tasty"
)

type Status struct {
	mu     sync.RWMutex
	fname  string
	states StrategyStatus
}

type StrategyStatus struct {
	SeqPfid    int                    `json:"seq-pfid"`
	Strategies map[string]stratOrders `json:"strategies"`
}

type stratOrders struct {
	LastSubmitted time.Time `json:"last-submitted"`
	//RetryConfig   RetryConfig            `json:"retry-config"`
	//OrderDetails map[string]orderDetail `json:"order-details"`
	WrappedOrders []WrappedOrder `json:"wrapped-orders"`
}

type WrappedOrder struct {
	PreflightID   string         `json:"pfid"`
	SubmitTime    time.Time      `json:"submit-time"`
	LastRetry     time.Time      `json:"last-retry"`
	RetryAttempts int            `json:"retry-attempts"`
	Order         tasty.NewOrder `json:"order"`
	// TODO: other submit metrics here?
}

type orderDetail struct {
	OrderId string  `json:"order-id"`
	Price   float64 `json:"price"`
	Status  string  `json:"status"`
}

func NewStatus(filename string) *Status {
	states := StrategyStatus{
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
	return &Status{
		states: states,
		fname:  filename,
	}

}

func newStratOrders(ts time.Time, pfid string, order tasty.NewOrder) stratOrders {
	return stratOrders{
		LastSubmitted: ts,
		WrappedOrders: []WrappedOrder{newWrappedOrder(ts, pfid, order)},
	}
}

func newWrappedOrder(ts time.Time, pfid string, order tasty.NewOrder) WrappedOrder {
	return WrappedOrder{
		PreflightID:   pfid,
		RetryAttempts: 0,
		Order:         order,
		SubmitTime:    ts,
	}
}

//func newStratOrders(ts time.Time, pfid string) stratOrders {
//	return stratOrders{
//		LastSubmitted: ts,
//		OrderDetails: map[string]orderDetail{
//			pfid: orderDetail{},
//		},
//	}
//}

func (ss *Status) PPrint() {
	bytes, _ := json.MarshalIndent(ss.states, "", "  ")
	fmt.Println("Strategy States:")
	fmt.Println(string(bytes))
}

func (ss *Status) Submit(stratname string, ts time.Time, pfid string, order tasty.NewOrder) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if orders, ok := ss.states.Strategies[stratname]; !ok {
		ss.states.Strategies[stratname] = newStratOrders(ts, pfid, order)
	} else {
		orders.LastSubmitted = ts
		//orders.OrderDetails[pfid] = orderDetail{}
		orders.WrappedOrders = append(orders.WrappedOrders, newWrappedOrder(ts, pfid, order))
		ss.states.Strategies[stratname] = orders
	}

	ss.writefile()
}

//func (ss *StratStates) Submit(stratname string, ts time.Time, pfid string) {
//	ss.mu.Lock()
//	defer ss.mu.Unlock()
//
//	if orders, ok := ss.states.Strategies[stratname]; !ok {
//		ss.states.Strategies[stratname] = newStratOrders(ts, pfid)
//	} else {
//		orders.LastSubmitted = ts
//		orders.OrderDetails[pfid] = orderDetail{}
//		ss.states.Strategies[stratname] = orders
//	}
//
//	ss.writefile()
//}

func (ss *Status) StatusByName(stratname string) (stratOrders, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if orders, ok := ss.states.Strategies[stratname]; ok {
		return orders, nil
	}
	return stratOrders{}, fmt.Errorf("No status for strategy name")
}

func (ss *Status) LastSubmitted(stratname string) (time.Time, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if state, ok := ss.states.Strategies[stratname]; ok {
		return state.LastSubmitted, nil
	}
	return time.Now().AddDate(-1, 0, 0), fmt.Errorf("No status for strategy name")
}

func (ss *Status) NextPFID() int {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.states.SeqPfid += 1
	ss.writefile()
	return ss.states.SeqPfid
}

func (ss *Status) UpdateByID(id string) {}

func (ss *Status) writefile() {
	byt, err := json.MarshalIndent(ss.states, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(ss.fname, byt, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
