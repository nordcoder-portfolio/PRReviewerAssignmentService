package app

import (
	"avito_test/config"
	httptransport "avito_test/internal/http"
	"avito_test/internal/http/health"
	"avito_test/internal/repo/postgres"

	prcontroller "avito_test/internal/controller/pr"
	statscontroller "avito_test/internal/controller/stats"
	teamcontroller "avito_test/internal/controller/team"
	usercontroller "avito_test/internal/controller/user"

	prusecase "avito_test/internal/usecase/pr"
	statsusecase "avito_test/internal/usecase/stats"
	teamusecase "avito_test/internal/usecase/team"
	userusecase "avito_test/internal/usecase/user"

	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg        config.Config
	logger     *slog.Logger
	db         *pgxpool.Pool
	httpServer *http.Server
}

func NewHTTPHandler(logger *slog.Logger, db *pgxpool.Pool) http.Handler {
	tx := postgres.NewTransactor(db)

	teamRepo := postgres.NewTeamRepository(tx)
	userRepo := postgres.NewUserRepository(tx)
	prRepo := postgres.NewPRRepository(tx)

	usecaseLogger := logger.With(slog.String("layer", "usecase"))
	teamLoggerUC := usecaseLogger.With(slog.String("module", "team"))
	userLoggerUC := usecaseLogger.With(slog.String("module", "user"))
	prLoggerUC := usecaseLogger.With(slog.String("module", "pr"))
	statsLoggerUC := usecaseLogger.With(slog.String("module", "stats"))

	// todo reduce amount of uc args
	chooser := prusecase.NewRandomReviewerChooser()
	teamUC := teamusecase.New(teamRepo, userRepo, prRepo, chooser, tx, teamLoggerUC)
	userUC := userusecase.New(userRepo, prRepo, tx, userLoggerUC)
	prUC := prusecase.New(userRepo, prRepo, chooser, tx, prLoggerUC)
	statsUC := statsusecase.New(prRepo, statsLoggerUC)

	ctrlLogger := logger.With(slog.String("layer", "controller"))

	teamLoggerCtrl := ctrlLogger.With(slog.String("module", "team"))
	userLoggerCtrl := ctrlLogger.With(slog.String("module", "user"))
	prLoggerCtrl := ctrlLogger.With(slog.String("module", "pr"))
	statsLoggerCtrl := ctrlLogger.With(slog.String("module", "stats"))

	teamCtrl := teamcontroller.New(teamUC, teamLoggerCtrl)
	userCtrl := usercontroller.New(userUC, userLoggerCtrl)
	prCtrl := prcontroller.New(prUC, prLoggerCtrl)
	statsCtrl := statscontroller.New(statsUC, statsLoggerCtrl)

	healthSvc := health.NewService(db)

	server := httptransport.NewServer(
		teamCtrl,
		userCtrl,
		prCtrl,
		statsCtrl,
		healthSvc,
		logger.With(slog.String("layer", "server")),
	)

	return httptransport.NewHandler(server)
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*App, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse db dsn: %w", err)
	}

	pgxCfg.MaxConns = cfg.DB.MaxConns
	pgxCfg.MinConns = cfg.DB.MinConns
	pgxCfg.MaxConnLifetime = cfg.DB.MaxConnLifetime
	pgxCfg.MaxConnIdleTime = cfg.DB.MaxConnIdleTime

	db, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}

	logger.Info("connected to postgres",
		slog.String("dsn", maskDSN(cfg.DB.DSN)),
		slog.Int("max_conns", int(cfg.DB.MaxConns)),
	)

	handler := NewHTTPHandler(logger, db)

	httpServer := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         db,
		httpServer: httpServer,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("starting http server",
		slog.String("addr", a.httpServer.Addr),
		slog.String("env", a.cfg.Env),
	)

	err := a.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server listen: %w", err)
	}

	a.logger.Info("http server stopped")
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down http server")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("http server shutdown failed", slog.String("error", err.Error()))
		a.db.Close()
		return fmt.Errorf("http server shutdown: %w", err)
	}

	a.logger.Info("closing db pool")
	a.db.Close()

	a.logger.Info("shutdown completed successfully")
	return nil
}

func maskDSN(dsn string) string {
	if len(dsn) > 0 {
		return "***"
	}
	return ""
}
