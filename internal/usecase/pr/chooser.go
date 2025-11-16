package pr

import (
	"math/rand"
	"time"
)

var _ ReviewerChooser = (*randomReviewerChooser)(nil)

type ReviewerChooser interface {
	Choice(candidates []string, limit int) []string
}

type randomReviewerChooser struct {
	rnd *rand.Rand
}

func NewRandomReviewerChooser() *randomReviewerChooser {
	return &randomReviewerChooser{
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *randomReviewerChooser) Choice(candidates []string, limit int) []string {
	n := len(candidates)
	if n == 0 || limit <= 0 {
		return []string{}
	}
	if limit >= n {
		out := make([]string, n)
		copy(out, candidates)
		return out
	}

	out := make([]string, 0, limit)
	indexes := make([]int, n)
	for i := 0; i < n; i++ {
		indexes[i] = i
	}

	for i := 0; i < limit; i++ {
		j := i + c.rnd.Intn(n-i)
		indexes[i], indexes[j] = indexes[j], indexes[i]
		out = append(out, candidates[indexes[i]])
	}

	return out
}
