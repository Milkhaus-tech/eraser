package browser

import "testing"

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
