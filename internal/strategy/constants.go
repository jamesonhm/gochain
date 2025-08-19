package strategy

type StrikeMethod string
type OptType string
type OptSide string

const (
	Delta    StrikeMethod = "delta"
	Relative StrikeMethod = "relative"
	//
	Call OptType = "C"
	Put  OptType = "P"
	//
	Buy  OptSide = "buy"
	Sell OptSide = "sell"
)
