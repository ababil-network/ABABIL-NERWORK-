package app

import "testing"

func TestNewAppKeepers(t *testing.T) {
	keepers := NewAppKeepers()

	if keepers == nil {
		t.Fatal("NewAppKeepers() returned nil")
	}
}
