package http

import (
	"avito_test/internal/api"
	"avito_test/internal/domain"
	"encoding/json"
	"net/http"
)

func (s *Server) writeError(w http.ResponseWriter, err error) {
	if appErr, ok := domain.AsAppError(err); ok {
		resp := api.ErrorResponse{}
		resp.Error.Code = api.ErrorResponseErrorCode(appErr.Code)
		resp.Error.Message = appErr.Msg

		s.writeJSON(w, appErr.Status, resp)
		return
	}

	internal := domain.Internal()

	resp := api.ErrorResponse{}
	resp.Error.Code = api.ErrorResponseErrorCode(internal.Code)
	resp.Error.Message = internal.Msg

	s.writeJSON(w, internal.Status, resp)
}

func (s *Server) writeBadRequest(w http.ResponseWriter, msg string) {
	s.writeError(w, domain.BadRequest(msg))
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		s.logger.Error("failed to encode response")
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(data)
	if err != nil {
		s.logger.Error("failed to write response")
		return
	}
}
