package digest

import (
	"fmt"
	"strings"
)

var ForbiddenPhrases = []string{
	"Personal Data Removal Request",
	"removal of my personal information from your database",
	"CCPA", "do not sell", "data subject", "opt-out", "opt out",
	"deletion request", "privacy request", "data request", "deletion of data",
	"delete personal", "your deletion", "removal request",
}

type Item struct {
	BrokerName string
	Kind       string
	URL        string
}

// Render builds the private monthly summary.
func Render(items []Item, confirmed, submitted int) (subject, body string) {
	subject = fmt.Sprintf("Eraser monthly: %d sites need you", len(items))
	var b strings.Builder
	fmt.Fprintf(&b, "Sites requiring your attention: %d\n\n", len(items))
	for _, item := range items {
		fmt.Fprintf(&b, "- %s (%s)\n  %s\n", item.BrokerName, item.Kind, item.URL)
	}
	fmt.Fprintf(&b, "\nAutomated this run: %d confirmed, %d submitted.\n", confirmed, submitted)
	return subject, b.String()
}
