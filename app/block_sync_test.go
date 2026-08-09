package app

import "testing"

func TestValidateBlockRequest(t *testing.T) {
	tests := []struct {
		name    string
		request BlockRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: BlockRequest{
				FromHeight: 1,
				ToHeight:   10,
			},
			valid: true,
		},
		{
			name: "invalid from height",
			request: BlockRequest{
				FromHeight: 0,
				ToHeight:   10,
			},
			valid: false,
		},
		{
			name: "invalid height range",
			request: BlockRequest{
				FromHeight: 10,
				ToHeight:   5,
			},
			valid: false,
		},
		{
			name: "too many blocks",
			request: BlockRequest{
				FromHeight: 1,
				ToHeight:   MaxBlocksPerSyncRequest + 1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBlockRequest(tt.request)

			if tt.valid && err != nil {
				t.Fatalf("expected valid request, got error: %v", err)
			}

			if !tt.valid && err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
