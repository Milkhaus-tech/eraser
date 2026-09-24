package digest

import (
	"strings"
	"testing"
)

func TestRenderPreservesItemTextAndKeepsFixedTextSafe(t *testing.T) {
	item := Item{BrokerName: "Example People Search", Kind: "manual step", URL: "https://example.com/opt-out?token=secret"}
	subject, body := Render([]Item{item}, 2, 3)
	if subject != "Eraser monthly: 1 sites need you" {
		t.Fatalf("subject = %q", subject)
	}
	if !strings.Contains(body, item.URL) {
		t.Errorf("body does not contain URL verbatim: %q", body)
	}
	fixedText := subject + "\nSites requiring your attention: 1\nAutomated this run: 2 confirmed, 3 submitted."
	for _, phrase := range ForbiddenPhrases {
		if strings.Contains(strings.ToLower(fixedText), strings.ToLower(phrase)) {
			t.Errorf("contains forbidden phrase %q", phrase)
		}
	}
}
