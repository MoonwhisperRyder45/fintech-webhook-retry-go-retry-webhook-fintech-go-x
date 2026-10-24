package payment

import "testing"

func TestDecideRiskAction(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{{"ordinary", 20, "deliver"}, {"high risk", 80, "hold"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := Decide(Event{ID: "pay-1", Amount: 100, Currency: "USD", RiskScore: tt.score})
			if err != nil || n.Action != tt.want {
				t.Fatalf("got %#v, %v", n, err)
			}
		})
	}
}
