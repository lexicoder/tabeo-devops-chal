package apiutils

import (
	"net/http"
)

func AllowedMethods(next http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		found := existsInSlice(methods, r.Method)
		if found {
			next(w, r)
		} else {
			RenderResponse(r, w, http.StatusMethodNotAllowed, nil)
		}
	}
}

func AllowedContentTypes(next http.HandlerFunc, mediaTypes ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		found := existsInSlice(mediaTypes, r.Header.Get("content-type"))
		if found {
			next(w, r)
		} else {
			RenderResponse(r, w, http.StatusUnsupportedMediaType, nil)
		}
	}
}

func existsInSlice(list []string, needle string) bool {
	for i := range list {
		if list[i] == needle {
			return true
		}
	}
	return false
}
