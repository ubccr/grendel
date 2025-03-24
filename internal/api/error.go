package api

import (
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
)

func ErrorHandler(err error) error {
	var errorStatus fuego.ErrorWithStatus
	switch {
	case errors.As(err, &fuego.HTTPError{}),
		errors.As(err, &errorStatus):
		return handleHTTPError(err)
	}

	return err
}

func handleHTTPError(err error) fuego.HTTPError {
	errResponse := fuego.HTTPError{
		Err: err,
	}

	var errorInfo fuego.HTTPError
	if errors.As(err, &errorInfo) {
		errResponse = errorInfo
	}

	// Check status code
	var errorStatus fuego.ErrorWithStatus
	if errors.As(err, &errorStatus) {
		errResponse.Status = errorStatus.StatusCode()
	}

	// Check for detail
	var errorDetail fuego.ErrorWithDetail
	if errors.As(err, &errorDetail) {
		errResponse.Detail = errorDetail.DetailMsg()
	}

	if errResponse.Title == "" {
		errResponse.Title = http.StatusText(errResponse.Status)
	}
	log.Errorf("%s, status=%d, detail=%s, error=%s", errResponse.Title, errResponse.StatusCode(), errResponse.DetailMsg(), errResponse.Err.Error())
	// slog.Error("Error "+errResponse.Title, "status", errResponse.StatusCode(), "detail", errResponse.DetailMsg(), "error", errResponse.Err)

	return errResponse
}

func ErrorSerializer(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	var errorStatus fuego.ErrorWithStatus
	if errors.As(err, &errorStatus) {
		status = errorStatus.StatusCode()
	}

	w.Header().Set("Content-Type", "application/json")

	// // ogen does not support application/problem+json content types
	// var httpError HTTPError
	// if errors.As(err, &httpError) {
	// 	w.Header().Set("Content-Type", "application/problem+json")
	// }

	w.WriteHeader(status)
	_ = fuego.SendJSON(w, nil, err)
}
