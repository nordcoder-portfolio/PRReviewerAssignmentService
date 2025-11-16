package domain

import (
	"errors"
	"net/http"
)

type ErrorCode string

const (
	CodeTeamExists  ErrorCode = "TEAM_EXISTS"
	CodePRExists    ErrorCode = "PR_EXISTS"
	CodePRMerged    ErrorCode = "PR_MERGED"
	CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	CodeNoCandidate ErrorCode = "NO_CANDIDATE"
	CodeNotFound    ErrorCode = "NOT_FOUND"

	CodeBadRequest ErrorCode = "BAD_REQUEST"
	CodeInternal   ErrorCode = "INTERNAL_ERROR"
)

type AppError struct {
	Code   ErrorCode
	Msg    string
	Status int
}

func (e *AppError) Error() string {
	return e.Msg
}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func NewAppError(code ErrorCode, msg string, status int) *AppError {
	return &AppError{
		Code:   code,
		Msg:    msg,
		Status: status,
	}
}

func BadRequest(msg string) *AppError {
	return NewAppError(CodeBadRequest, msg, http.StatusBadRequest)
}

func NotFound(msg string) *AppError {
	return NewAppError(CodeNotFound, msg, http.StatusNotFound)
}

func TeamExists(msg string) *AppError {
	return NewAppError(CodeTeamExists, msg, http.StatusBadRequest)
}

func PRExists(msg string) *AppError {
	return NewAppError(CodePRExists, msg, http.StatusConflict)
}

func PRMerged(msg string) *AppError {
	return NewAppError(CodePRMerged, msg, http.StatusConflict)
}

func NotAssigned(msg string) *AppError {
	return NewAppError(CodeNotAssigned, msg, http.StatusConflict)
}

func NoCandidate(msg string) *AppError {
	return NewAppError(CodeNoCandidate, msg, http.StatusConflict)
}

func Internal() *AppError {
	return &AppError{
		Code:   CodeInternal,
		Msg:    "internal error",
		Status: http.StatusInternalServerError,
	}
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
