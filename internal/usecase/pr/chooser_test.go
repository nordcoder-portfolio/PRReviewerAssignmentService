package pr

import (
	"slices"
	"testing"
)

func TestRandomReviewerChooser_LimitTwo(t *testing.T) {
	t.Parallel()

	chooser := NewRandomReviewerChooser()

	tests := []struct {
		name       string
		candidates []string
		limit      int
		wantLen    int
		wantAll    bool
	}{
		{
			name:       "0 - 0",
			candidates: []string{},
			limit:      2,
			wantLen:    0,
			wantAll:    false,
		},
		{
			name:       "2 - 1",
			candidates: []string{"u1"},
			limit:      2,
			wantLen:    1,
			wantAll:    true,
		},
		{
			name:       "2 - 2",
			candidates: []string{"u1", "u2"},
			limit:      2,
			wantLen:    2,
			wantAll:    true,
		},
		{
			name:       "3 - 2",
			candidates: []string{"u1", "u2", "u3"},
			limit:      2,
			wantLen:    2,
			wantAll:    false,
		},
		{
			name:       "5 - 2",
			candidates: []string{"u1", "u2", "u3", "u4", "u5"},
			limit:      2,
			wantLen:    2,
			wantAll:    false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := chooser.Choice(tt.candidates, tt.limit)

			if len(got) != tt.wantLen {
				t.Fatalf("Choice(%v, %d) len = %d, want %d",
					tt.candidates, tt.limit, len(got), tt.wantLen)
			}

			if tt.wantAll {
				if len(tt.candidates) != len(got) {
					t.Fatalf("wantAll=true but len(candidates)=%d, len(got)=%d",
						len(tt.candidates), len(got))
				}
				for _, c := range tt.candidates {
					if !slices.Contains(got, c) {
						t.Fatalf("result %v does not contain candidate %q", got, c)
					}
				}
			}

			if len(got) > 1 {
				seen := make(map[string]struct{}, len(got))
				for _, id := range got {
					if _, ok := seen[id]; ok {
						t.Fatalf("result %v contains duplicate id %q", got, id)
					}
					seen[id] = struct{}{}
					if !slices.Contains(tt.candidates, id) {
						t.Fatalf("result %v contains id %q not in candidates %v", got, id, tt.candidates)
					}
				}
			}
		})
	}
}
