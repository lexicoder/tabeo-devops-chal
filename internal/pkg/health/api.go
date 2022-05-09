package health

import (
	"net/http"
)

func HealthGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
