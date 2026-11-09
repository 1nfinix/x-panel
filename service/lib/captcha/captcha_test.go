package captcha

import (
	"strings"
	"testing"
	"time"
)

func TestExpiringStore(t *testing.T) {
	now := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)
	store := NewExpiringStore()
	store.now = func() time.Time { return now }
	if err := store.Set("one", "Ab12"); err != nil {
		t.Fatal(err)
	}
	if !store.Verify("one", "aB12", true) || store.Verify("one", "Ab12", true) {
		t.Fatal("captcha must be case insensitive and single use")
	}
	store.Set("wrong", "Ab12")
	if store.Verify("wrong", "bad", true) || store.Verify("wrong", "Ab12", true) {
		t.Fatal("wrong answer must consume captcha")
	}
	store.Set("expired", "Ab12")
	now = now.Add(5 * time.Minute)
	if store.Verify("expired", "Ab12", true) {
		t.Fatal("captcha must expire within five minutes")
	}
}

func TestGeneratedAnswerHasLettersAndDigits(t *testing.T) {
	for i := 0; i < 20; i++ {
		id := string(rune('a' + i))
		if _, err := GenerateCaptchaHandler(id, 160, 48); err != nil {
			t.Fatal(err)
		}
		answer := Store.Get(id, true)
		if len(answer) != 4 || !strings.ContainsAny(answer, "23456789") || !strings.ContainsAny(answer, "ABCDEFGHJKLMNPQRSTUVWXYZ") {
			t.Fatalf("invalid generated answer %q", answer)
		}
	}
}
