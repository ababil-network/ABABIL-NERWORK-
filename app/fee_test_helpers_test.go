package app

import (
	"testing"
	"time"
)

// setupValidFeeEnvironment initializes the deterministic fee environment
// required by transaction/state-transition tests.
//
// Production transaction creation MUST still use the validated reference
// price. This helper only provides a valid test fixture.
func setupValidFeeEnvironment(t *testing.T) {
	t.Helper()

	NodeReferencePrice.Reset()

	// Test reference price:
	// 1 ABABIL = $1.00
	// Therefore 1 ABABIL = 1,000,000 micro-USD.
	price := uint64(1_000_000)

	if err := NodeReferencePrice.AddObservation(
		price,
		time.Now().UTC(),
	); err != nil {
		t.Fatalf("failed to initialize ABABIL reference price: %v", err)
	}

	if err := NodeDynamicFee.SetLoadPercent(0); err != nil {
		t.Fatalf("failed to initialize network load: %v", err)
	}
}
