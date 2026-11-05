package obs

import (
	"strings"
	"testing"
)

func TestKnownSeriesStartAtZero(t *testing.T) {
	out := render(NewMetrics().Reg)
	for _, want := range []string{
		`consumer_messages_total{outcome="dlq"} 0`,
		`consumer_messages_total{outcome="processed"} 0`,
		`claims_submitted_total{result="created"} 0`,
		`authz_denied_total{resource="claim"} 0`,
		`outbox_publish_total{result="error"} 0`,
		"http_panics_total 0",
	} {
		if !strings.Contains(out, want+"\n") {
			t.Errorf("missing %q", want)
		}
	}
}
