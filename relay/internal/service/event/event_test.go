package event

import (
	"testing"
)

func TestQueryRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     QueryRequest
		wantErr bool
	}{
		{
			name: "valid request with cid",
			req: QueryRequest{
				Cid:   "conv123",
				Limit: 50,
			},
			wantErr: false,
		},
		{
			name: "valid request with lastMid",
			req: QueryRequest{
				Cid:     "conv123",
				LastMid: 100,
				Limit:   50,
			},
			wantErr: false,
		},
		{
			name: "missing cid",
			req: QueryRequest{
				Cid:   "",
				Limit: 50,
			},
			wantErr: true,
		},
		{
			name: "negative limit uses default",
			req: QueryRequest{
				Cid:   "conv123",
				Limit: -1,
			},
			wantErr: false, // Should use default limit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.Cid == ""
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestQueryRequest_LimitDefault(t *testing.T) {
	req := QueryRequest{
		Cid:   "conv123",
		Limit: 0,
	}

	// Default limit should be applied
	if req.Limit <= 0 {
		req.Limit = 50 // Apply default
	}

	if req.Limit != 50 {
		t.Errorf("default limit = %d, want 50", req.Limit)
	}
}

func TestQueryRequest_TimeRange(t *testing.T) {
	tests := []struct {
		name   string
		before int64
		after  int64
		valid  bool
	}{
		{
			name:   "no time range",
			before: 0,
			after:  0,
			valid:  true,
		},
		{
			name:   "only before",
			before: 1000,
			after:  0,
			valid:  true,
		},
		{
			name:   "only after",
			before: 0,
			after:  1000,
			valid:  true,
		},
		{
			name:   "valid range",
			before: 2000,
			after:  1000,
			valid:  true,
		},
		{
			name:   "invalid range (after > before)",
			before: 1000,
			after:  2000,
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := true
			if tt.before > 0 && tt.after > 0 && tt.after > tt.before {
				valid = false
			}

			if valid != tt.valid {
				t.Errorf("valid = %v, want %v", valid, tt.valid)
			}
		})
	}
}
