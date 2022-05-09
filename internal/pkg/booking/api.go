package booking

import (
	"net/http"
	"strconv"

	"spacetrouble/pkg/apiutils"
)

func BookingHandler(srv BookingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			create(srv, w, r)
		} else if r.Method == http.MethodGet {
			all(srv, w, r)
		}
	}
}

func create(srv BookingService, w http.ResponseWriter, r *http.Request) {
	var bookReq BookingRequest
	if err := apiutils.JsonDecodeBody(r, &bookReq); err != nil {
		ae := apiutils.NewBadRequest("error json decoding body")
		apiutils.RenderResponse(r, w, ae.StatusCode, ae)
		return
	}

	if err := bookReq.Validate(); err != nil {
		ae := apiutils.NewBadRequest(err.Error())
		apiutils.RenderResponse(r, w, ae.StatusCode, ae)
		return
	}

	ans, err := srv.MakeBooking(r.Context(), bookReq)
	if err != nil {
		ae := getApiError(err)
		apiutils.RenderResponse(r, w, ae.StatusCode, ae)
		return

	}
	apiutils.RenderResponse(r, w, http.StatusCreated, ans)
}

func all(srv BookingService, w http.ResponseWriter, r *http.Request) {
	var limit int
	if keys, ok := r.URL.Query()["limit"]; ok {
		if len(keys) > 0 && len(keys[0]) > 0 {
			var err error
			limit, err = strconv.Atoi(keys[0])
			if err != nil {
				ae := apiutils.NewBadRequest(err.Error())
				apiutils.RenderResponse(r, w, ae.StatusCode, ae)
				return
			}
			if limit < 0 {
				ae := apiutils.NewBadRequest("negative limit")
				apiutils.RenderResponse(r, w, ae.StatusCode, ae)
				return
			}
		}
	}
	var cur string
	if keys, ok := r.URL.Query()["cursor"]; ok {
		if len(keys) > 0 && len(keys[0]) > 0 {
			cur = keys[0]
		}
	}
	if limit == 0 {
		limit = 10
	}
	getReq := GetBookingsReq{
		Limit: limit,
	}
	if len(cur) > 0 {
		var err error
		getReq.Ts, getReq.Uuid, err = decodeCursor(cur)
		if err != nil {
			ae := apiutils.NewBadRequest(err.Error())
			apiutils.RenderResponse(r, w, ae.StatusCode, ae)
			return
		}
	}

	ans, err := srv.AllBookings(r.Context(), getReq)
	if err != nil {
		ae := getApiError(err)
		apiutils.RenderResponse(r, w, ae.StatusCode, ae)
		return
	}

	apiutils.RenderResponse(r, w, http.StatusOK, ans)
}

func getApiError(err error) apiutils.ApiError {
	ae := apiutils.ApiError{Msg: err.Error()}
	switch err {
	case ErrInvalidUUID:
		ae.StatusCode = http.StatusBadRequest
	case ErrMissingDestination:
		ae.StatusCode = http.StatusNotFound
	case ErrLaunchPadUnavailable:
		ae.StatusCode = http.StatusConflict
	default:
		ae.StatusCode = http.StatusInternalServerError
	}
	return ae
}
