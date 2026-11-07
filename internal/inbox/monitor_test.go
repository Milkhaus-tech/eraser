package inbox

import (
	"strings"
	"testing"

	"github.com/eraser-privacy/eraser/internal/broker"
	"github.com/eraser-privacy/eraser/internal/config"
)

func TestSharedMailHostMatchesOnlyTheExactAddress(t *testing.T) {
	m := NewMonitor(config.InboxConfig{}, []broker.Broker{
		{ID: "free", Email: "Skip.Smasher@gmail.com", Website: "https://free.example"},
		{ID: "real", Email: "privacy@real.example", Website: "https://www.real.example/about"},
		{ID: "hosted", Email: "x@y.example", Website: "https://docs.google.com/about"},
	})
	cases := []struct {
		address, domain, want string
	}{
		{"skip.smasher@gmail.com", "gmail.com", "free"},
		{"attacker@gmail.com", "gmail.com", ""},
		{"anyone@docs.google.com", "docs.google.com", ""},
		{"dpo@real.example", "real.example", "real"},
		{"anyone@free.example", "free.example", "free"},
		{"attacker@evil.example", "evil.example", ""},
	}
	for _, c := range cases {
		b, ok := m.matchBroker(c.address, c.domain)
		if ok != (c.want != "") || b.ID != c.want {
			t.Errorf("%s: got %q ok=%v, want %q", c.address, b.ID, ok, c.want)
		}
	}
}

func TestDmarcPassed(t *testing.T) {
	gmail := "Authentication-Results: mx.google.com;\r\n       dkim=pass header.i=@real.example header.s=s1;\r\n" +
		"       spf=pass (google.com: domain of a@real.example designates 1.2.3.4) smtp.mailfrom=a@real.example;\r\n"
	message := func(headers ...string) []byte {
		return []byte(strings.Join(headers, "") + "From: a@real.example\r\nSubject: x\r\n\r\nbody\r\n")
	}
	cases := []struct {
		name string
		raw  []byte
		want bool
	}{
		{"pass", message(gmail + "       dmarc=pass (p=REJECT sp=REJECT dis=NONE) header.from=real.example\r\n"), true},
		{"case", message(gmail + "       dmarc=PASS (p=NONE) header.from=REAL.example\r\n"), true},
		{"fail", message(gmail + "       dmarc=fail (p=NONE) header.from=real.example\r\n"), false},
		{"none", message(gmail + "       dmarc=none header.from=real.example\r\n"), false},
		{"no dmarc clause", message("Authentication-Results: mx.google.com; spf=pass smtp.mailfrom=a@real.example\r\n"), false},
		{"no header", message(), false},
		{"other domain", message(gmail + "       dmarc=pass (p=NONE) header.from=evil.example\r\n"), false},
		{"other server", message("Authentication-Results: mx.evil.example; dmarc=pass header.from=real.example\r\n"), false},
		{"forged below the real one", message(
			gmail+"       dmarc=fail (p=NONE) header.from=real.example\r\n",
			"Authentication-Results: mx.google.com; dmarc=pass header.from=real.example\r\n"), false},
		{"pass hidden in a comment", message(gmail + "       dmarc=fail (dmarc=pass header.from=real.example) header.from=real.example\r\n"), false},
		{"pass planted in an earlier comment", message("Authentication-Results: mx.google.com; " +
			"spf=pass (dmarc=pass header.from=real.example ) smtp.mailfrom=a@real.example; dmarc=fail header.from=real.example\r\n"), false},
		{"not a message", []byte("garbage"), false},
	}
	for _, c := range cases {
		if got := dmarcPassed(c.raw, "real.example"); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestGuardActions(t *testing.T) {
	links := func() ClassifiedResponse {
		return ClassifiedResponse{FormURL: "https://real.example/form", ConfirmURL: "https://real.example/c", Reason: "x"}
	}
	c := links()
	GuardActions(&Email{}, &c)
	if c.FormURL != "" || c.ConfirmURL != "" || !c.NeedsReview || !strings.Contains(c.Reason, "DMARC") {
		t.Errorf("an unproven sender kept its links: %+v", c)
	}
	c = links()
	GuardActions(&Email{Authenticated: true}, &c)
	if c.FormURL == "" || c.ConfirmURL == "" || c.NeedsReview {
		t.Errorf("a proven sender lost its links: %+v", c)
	}
	c = ClassifiedResponse{}
	GuardActions(&Email{}, &c)
	if c.NeedsReview {
		t.Error("a reply with no links needs no review for this")
	}
}
