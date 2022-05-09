package apiutils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ApiError struct {
	StatusCode int    `json:"-"`
	Msg        string `json:"error,omitempty"`
}

func (o *ApiError) Error() string {
	return fmt.Sprintf("%d: %s", o.StatusCode, o.Msg)
}

func NewInternalServerError(msg string) ApiError {
	return ApiError{http.StatusInternalServerError, msg}
}

func NewBadRequest(msg string) ApiError {
	return ApiError{http.StatusBadRequest, msg}
}

func JsonDecodeBody(r *http.Request, dst interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}

func RenderResponse(r *http.Request, w http.ResponseWriter, statusCode int, res interface{}) {
	// TODO check request headers to determine the response type
	// for the task only json supported
	contentType := r.Header.Get("Content-type")
	switch contentType {
	case "application/json":
		renderJson(w, statusCode, res)
	default:
		renderJson(w, http.StatusUnsupportedMediaType, nil)
	}
}

func renderJson(w http.ResponseWriter, statusCode int, res interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var body []byte
	if res != nil {
		var err error
		body, err = json.Marshal(res)
		if err != nil {
			ae := NewInternalServerError(err.Error())
			statusCode = ae.StatusCode
			body, err = json.Marshal(&ae)
			if err != nil {
				body = []byte(`{"msg": "` + err.Error() + `"}`)
			}
		}
	}
	w.WriteHeader(statusCode)
	if len(body) > 0 {
		w.Write(body)
	}
}
