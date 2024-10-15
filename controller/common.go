package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/namlh/vulcanLabsOA/consts/errcode"
)

type Validator interface {
	// Valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	Valid(ctx context.Context) map[string]string
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	buf, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	if _, err = w.Write(buf); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func decodeValid[T Validator](r *http.Request) (T, error) {
	var v T

	buf := bytes.Buffer{}
	if _, err := io.Copy(&buf, r.Body); err != nil {
		return v, fmt.Errorf("copy to buffer: %w", err)
	}

	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		if syntaxErr := (*json.SyntaxError)(nil); errors.As(err, &syntaxErr) {
			return v, AppError{
				ErrCode:    errcode.InvalidParameters,
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid json syntax",
				err:        err,
			}
		}
		return v, fmt.Errorf("decode json: %w", err)
	}

	if err := r.Body.Close(); err != nil {
		return v, fmt.Errorf("close body: %w", err)
	}

	problems := v.Valid(r.Context())
	if len(problems) > 0 {
		return v, AppError{
			ErrCode:    errcode.InvalidParameters,
			HttpStatus: http.StatusUnprocessableEntity,
			err:        ValidationErrors(problems),
		}
	}

	return v, nil
}

type AppError struct {
	ErrCode    int
	Message    string
	HttpStatus int
	err        error
}

func (e AppError) Error() string {
	return e.err.Error()
}

type ErrResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type SuccessResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

func NewSuccessResponse[T any](data T) SuccessResponse[T] {
	return SuccessResponse[T]{
		Code:    errcode.Success,
		Message: errcode.Text(errcode.Success),
		Data:    data,
	}
}

type ValidationErrors map[string]string

func (e ValidationErrors) Error() string {
	var err error
	for field, msg := range e {
		err = errors.Join(err, errors.New(field+":"+msg))
	}
	if err != nil {
		return err.Error()
	}

	return ""
}

func handleError(
	ctx context.Context,
	logger *slog.Logger,
	msg string,
	w http.ResponseWriter,
	err error,
) {
	logger.ErrorContext(ctx, msg, "error", err)

	if appErr := (AppError{}); errors.As(err, &appErr) {
		resp := ErrResponse{
			Code:    appErr.ErrCode,
			Message: appErr.Message,
		}
		if resp.Message == "" {
			resp.Message = errcode.Text(appErr.ErrCode)
		}

		if vErr := (ValidationErrors{}); errors.As(appErr.err, &vErr) {
			resp.Details = vErr
		}

		if err = encode(w, appErr.HttpStatus, resp); err != nil {
			logger.ErrorContext(ctx, "encode failed", "error", err)
			internalErrResponse(w)
		}

		return
	}

	internalErrResponse(w)
}

func internalErrResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func easyHandler[T any](
	name string,
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	f func(ctx context.Context) (T, error),
) {
	var err error

	ctx := r.Context()

	data, err := f(ctx)
	if err != nil {
		handleError(ctx, logger, name, w, err)
		return
	}

	if err = encode(w, http.StatusOK, NewSuccessResponse(data)); err != nil {
		handleError(ctx, logger, name, w, err)
	}
}
