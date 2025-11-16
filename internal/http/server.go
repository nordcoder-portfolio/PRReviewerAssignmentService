package http

import (
	"avito_test/internal/api"
	"avito_test/internal/controller/pr"
	"avito_test/internal/controller/stats"
	"avito_test/internal/controller/team"

	"avito_test/internal/controller/user"
	"avito_test/internal/http/health"
	"avito_test/internal/http/middleware"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	teamCtrl  team.Controller
	userCtrl  user.Controller
	prCtrl    pr.Controller
	statsCtrl stats.Controller
	healthSvc *health.Service
	logger    *slog.Logger
}

func NewServer(
	teamCtrl team.Controller,
	userCtrl user.Controller,
	prCtrl pr.Controller,
	statsCtrl stats.Controller,
	health *health.Service,
	logger *slog.Logger,
) *Server {
	return &Server{
		teamCtrl:  teamCtrl,
		userCtrl:  userCtrl,
		prCtrl:    prCtrl,
		statsCtrl: statsCtrl,
		healthSvc: health,
		logger:    logger,
	}
}

func NewHandler(s *Server) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID) // todo put request id in logs
	r.Use(chimw.Recoverer)

	r.Use(middleware.LoggingMiddleware(s.logger))

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		reqID := chimw.GetReqID(r.Context())

		s.logger.With(
			slog.String("layer", "http"),
			slog.String("op", "request_bind_error"),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("request_id", reqID),
		).Debug("request validation/binding error",
			slog.String("error", err.Error()),
		)

		s.writeBadRequest(w, err.Error())
	}

	r.Get("/health", s.HealthHandler)

	return api.HandlerWithOptions(s, api.ChiServerOptions{
		BaseRouter:       r,
		ErrorHandlerFunc: errorHandler,
	})
}

const healthTimeout = 3 * time.Second

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()

	var res interface{} = map[string]string{"status": "ok"}
	statusCode := http.StatusOK

	if s.healthSvc != nil {
		result := s.healthSvc.Check(ctx)
		res = result

		if result.Status != health.StatusOK {
			statusCode = http.StatusServiceUnavailable
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.logger.Error("failed to encode health response", slog.Any("err", err))
	}
}
