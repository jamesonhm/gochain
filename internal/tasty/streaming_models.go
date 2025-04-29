package tasty

type QuoteStreamerTokenResult struct {
	QuoteStreamerToken QuoteStreamerToken `json:"data"`
}

// Response from the API quote streamer request.
type QuoteStreamerToken struct {
	// API quote token unique to the customer identified by the session
	// Quote streamer tokens are valid for 24 hours.
	Token     string `json:"token"`
	DXLinkURL string `json:"dxlink-url"`
	Level     string `json:"level"`
}
