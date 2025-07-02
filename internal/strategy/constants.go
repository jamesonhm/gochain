package strategy

type StrikeMethod string
type OptType string
type OptSide string

const (
	Delta  StrikeMethod = "Delta"
	Offset StrikeMethod = "Offset"
	//
	Call OptType = "Call"
	Put  OptType = "Put"
	//
	Buy  OptSide = "Buy"
	Sell OptSide = "Sell"
)
