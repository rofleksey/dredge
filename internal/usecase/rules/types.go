package rules

// EvalPayload carries data for middleware and templates.
type EvalPayload struct {
	Event       string
	Channel     string
	Username    string
	Text        string
	Title       string
	IntervalSec float64
}
