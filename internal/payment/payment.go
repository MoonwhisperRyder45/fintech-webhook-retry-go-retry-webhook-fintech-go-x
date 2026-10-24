package payment

import "fmt"

type Event struct {
	ID        string `json:"id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	RiskScore int    `json:"risk_score"`
}
type Notification struct {
	Event  Event  `json:"event"`
	Action string `json:"action"`
	Audit  string `json:"audit"`
}

func Decide(e Event) (Notification, error) {
	if e.ID == "" || e.Amount <= 0 || e.Currency == "" {
		return Notification{}, fmt.Errorf("invalid payment event")
	}
	action := "deliver"
	if e.RiskScore >= 80 {
		action = "hold"
	}
	return Notification{Event: e, Action: action, Audit: fmt.Sprintf("payment=%s risk=%d action=%s", e.ID, e.RiskScore, action)}, nil
}
