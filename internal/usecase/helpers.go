package helpers

import (
	"avito_test/internal/domain"
	"log/slog"
)

func ConvertAndLogError(log *slog.Logger, msg string, err error) error {
	if appErr, ok := domain.AsAppError(err); ok {
		return appErr
	}

	log.Error("failed to "+msg, slog.Any("err", err))
	return err
}
