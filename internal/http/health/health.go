package health

import (
	"context"
	"time"
)

const pingTimeout = 2 * time.Second

type DBChecker interface {
	Ping(ctx context.Context) error
}

type Service struct {
	db DBChecker
}

func NewService(db DBChecker) *Service {
	return &Service{db: db}
}

type Status string

const (
	StatusOK       Status = "ok"
	StatusDegraded Status = "degraded"
)

type Result struct {
	Status Status            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

func (s *Service) Check(ctx context.Context) Result {
	res := Result{
		Status: StatusOK,
		Checks: make(map[string]string),
	}

	if s.db != nil {
		hctx, cancel := context.WithTimeout(ctx, pingTimeout)
		defer cancel()

		if err := s.db.Ping(hctx); err != nil {
			res.Status = StatusDegraded
			res.Checks["db"] = err.Error()
		} else {
			res.Checks["db"] = "ok"
		}
	}

	return res
}
