package browser

import (
	"testing"

	"github.com/eraser-privacy/eraser/internal/broker"
)

func TestValidateDomainBoundaries(t *testing.T) {
	h := NewConfirmationHandler([]string{"spokeo.com"})
	for _, raw := range []string{"https://spokeo.com/form", "https://privacy.spokeo.com/form"} {
		valid, _, err := h.ValidateDomain(raw)
		if err != nil || !valid {
			t.Errorf("expected %q accepted: valid=%v err=%v", raw, valid, err)
		}
	}
	for _, raw := range []string{"https://evil.io/form", "https://spokeo.com.evil.io/form"} {
		valid, _, err := h.ValidateDomain(raw)
		if err != nil || valid {
			t.Errorf("expected %q rejected: valid=%v err=%v", raw, valid, err)
		}
	}
}

func TestSharedHostsAreNotAnchors(t *testing.T) {
	brokers := []broker.Broker{
		{ID: "gmailer", Email: "someone@gmail.com", Website: "https://gmailer.example", OptOutURL: "https://docs.google.com/forms/d/e/REAL/viewform"},
		{ID: "otherform", Email: "privacy@other.example", OptOutURL: "https://forms.gle/REALSHORT"},
		{ID: "desk", Email: "dpo@desk.example", Website: "https://www.zendesk.com", OptOutURL: "https://techtarget.zendesk.com/hc/requests/new"},
		{ID: "portal", Email: "dpo@portal.example", OptOutURL: "https://privacyportal.onetrust.com/webform/abc/def"},
	}
	for _, d := range BrokerDomains(brokers) {
		if broker.IsSharedHost(d) {
			t.Errorf("shared host %q became a trust anchor", d)
		}
	}
	h := NewBrokerHandler(brokers)

	for _, raw := range []string{
		"https://gmail.com/x", "https://mail.google.com/x",
		"https://docs.google.com/forms/d/e/ATTACKER/viewform",
		"https://forms.gle/ATTACKER",
		"https://attacker.zendesk.com/hc/requests/new",
		"https://techtarget.zendesk.com/hc/requests/other",
		"https://privacyportal.onetrust.com/webform/evil/form",
	} {
		if valid, _, err := h.ValidateDomain(raw); err != nil || valid {
			t.Errorf("expected %q rejected: valid=%v err=%v", raw, valid, err)
		}
	}
	for _, raw := range []string{
		"https://docs.google.com/forms/d/e/REAL/viewform",
		"https://docs.google.com/forms/d/e/REAL/viewform/?usp=sf_link",
		"https://forms.gle/REALSHORT",
		"https://techtarget.zendesk.com/hc/requests/new",
		"https://gmailer.example/optout",
		"https://other.example/x",
	} {
		if valid, _, err := h.ValidateDomain(raw); err != nil || !valid {
			t.Errorf("expected %q accepted: valid=%v err=%v", raw, valid, err)
		}
	}

	real := "https://docs.google.com/forms/d/e/REAL/viewform"
	if ok, _, _ := h.ValidateForBroker(real, "gmailer"); !ok {
		t.Error("a broker's own published form must pass")
	}
	if ok, _, _ := h.ValidateForBroker(real, "desk"); ok {
		t.Error("another broker's shared form must not pass")
	}
	if ok, _, _ := h.ValidateForBroker("https://other.example/x", "desk"); !ok {
		t.Error("an own-domain host is not tied to one broker record")
	}
	if ok, _, _ := h.ValidateForBroker("https://evil.example/x", "desk"); ok {
		t.Error("an unknown host must not pass")
	}
	if ok, _, _ := h.ValidateForBroker(real, ""); !ok {
		t.Error("a hand-given URL has no broker to bind to")
	}
}
