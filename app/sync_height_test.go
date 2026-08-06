package app

import "testing"

func TestNeedSync(t *testing.T) {

	if !NeedSync(100, 110) {
		t.Fatal("expected sync to be required")
	}

	if NeedSync(110, 100) {
		t.Fatal("unexpected sync required")
	}

	if NeedSync(100, 100) {
		t.Fatal("unexpected sync required")
	}
}
