package app

import "testing"

func TestFinalFeeLoadCurve(t *testing.T) {
	tests := []struct {
		name string
		load uint64
		want uint64
	}{
		{"zero load", 0, 50000},
		{"50 percent", 50, 50000},
		{"51 percent", 51, 49143},
		{"85 percent", 85, 20000},
		{"95 percent", 95, 10000},
		{"100 percent", 100, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NodeDynamicFee.SetLoadPercent(tt.load); err != nil {
				t.Fatalf("SetLoadPercent failed: %v", err)
			}

			got := NodeDynamicFee.TargetTransactionsPerUSD()

			if got != tt.want {
				t.Fatalf(
					"load=%d: got %d TX/USD, want %d",
					tt.load,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestFinalFeeCurveNeverIncreasesTXPerUSDWithLoad(t *testing.T) {
	var previous uint64

	for load := uint64(0); load <= 100; load++ {
		if err := NodeDynamicFee.SetLoadPercent(load); err != nil {
			t.Fatalf("load=%d: %v", load, err)
		}

		current := NodeDynamicFee.TargetTransactionsPerUSD()

		if load > 0 && current > previous {
			t.Fatalf(
				"fee curve moved in wrong direction at load=%d: previous=%d current=%d",
				load,
				previous,
				current,
			)
		}

		previous = current
	}
}
