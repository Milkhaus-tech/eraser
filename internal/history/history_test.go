package history

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestAddBrokerResponseDeduplicates(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	received := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	for i := 0; i < 2; i++ {
		if err := store.AddBrokerResponse(&BrokerResponse{BrokerID: "broker", BrokerName: "Broker", EmailFrom: "reply@example.com", EmailSubject: "Reply", ResponseType: "pending", ReceivedAt: received}); err != nil {
			t.Fatal(err)
		}
	}
	responses, err := store.GetBrokerResponses("", false, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(responses) != 1 {
		t.Fatalf("got %d responses, want 1", len(responses))
	}
}

func TestMarkBrokerResponsesReported(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var ids []int64
	for i, status := range []string{"failed", "skipped"} {
		response := &BrokerResponse{
			BrokerID: fmt.Sprintf("broker-%d", i), BrokerName: "Broker",
			EmailFrom: "reply@example.com", EmailSubject: fmt.Sprintf("Reply %d", i),
			ResponseType: "form_required", FormURL: "https://example.com/form",
			ReceivedAt: time.Date(2026, 1, 2, 3, 4, 5+i, 0, time.UTC),
		}
		if err := store.AddBrokerResponse(response); err != nil {
			t.Fatal(err)
		}
		if err := store.SetBrokerResponseAction(response.ID, status); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, response.ID)
	}
	if err := store.MarkBrokerResponsesReported(ids); err != nil {
		t.Fatal(err)
	}
	responses, err := store.GetFailedActionResponses(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(responses) != 0 {
		t.Fatalf("got %d reportable responses after marking reported, want 0", len(responses))
	}
}
