package tui

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestParseDelay(t *testing.T) {
	for _, value := range []string{"0", "15", " 45 ", "10080"} {
		if _, err := parseDelay(value); err != nil {
			t.Fatal(value, err)
		}
	}
	for _, value := range []string{"", "-1", "10081", "1.5", "99999999999999999999999", "soon"} {
		if _, err := parseDelay(value); err == nil {
			t.Fatal("accepted", value)
		}
	}
}
func TestDelayEscapeKeepsCurrentValue(t *testing.T) {
	scriptedForms(t, "\x1b")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	value, err := chooseDelay(ctx, 15)
	if value != 15 || !errors.Is(err, errBack) {
		t.Fatal(value, err)
	}
}
