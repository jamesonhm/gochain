package strategy

import (
	"sync"
	"time"
)

type StratStates struct {
	states map[string]*state
	mu     sync.RWMutex
}

type state struct {
	name          string
	lastSubmitted time.Time
	orderInfo     orderInfo
}

type orderInfo struct {
	id     string
	status string
}

func (ss *StratStates) Update(name string) {}
