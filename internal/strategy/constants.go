package strategy

type StrikeMethod string
type OptType string
type OptSide string

const (
	Delta  StrikeMethod = "delta"
	Offset StrikeMethod = "offset"
	//
	Call OptType = "call"
	Put  OptType = "put"
	//
	Buy  OptSide = "buy"
	Sell OptSide = "sell"
)
