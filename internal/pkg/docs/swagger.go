// Space Booking API - service to book a travel to space
//
//    provided by:
//	  Georgios Komninos
//
//     Schemes: http
//     Version: 1
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
package docs

import (
	"spacetrouble/internal/pkg/booking"
	//"spacetrouble/pkg/apiutils"
)

// swagger:route GET /v1/health Health Health
//
// Just returns 200 when service is up.
//
// ---
// produces:
// - application/json
// responses:
//   200:

// swagger:route GET /v1/bookings Bookings All
// Fetches all bookings.
// Supports pagination via the cursor query parameter
// ---
// produces:
//  - application/json
// responses:
// 	200: AllBookingsPaginated
//	400:
//  500:

// swagger:parameters All
type BookingAllParamsWrapper struct {
	// in:query
	Limit int `json:"limit"`
	// in:query
	Cursor string `json:"cursor"`
}

// OK
// The operation was processed successfully
//
// An AllBookingsResponse obj
// swagger:response AllBookingsPaginated
type AllBookingsPaginated struct {
	Limit    int                      `json:"limit"`
	Cursor   string                   `json:"cursor"`
	Bookings []BookingSuccessResponse `json:"bookings"`
}

type AllBookingsPaginatedResp struct {
}

// swagger:route POST /v1/bookings Bookings Booking
// Attempts to make a new booking to a space destination.
// ---
// produces:
// - application/json
// responses:
// 201: BookingSuccessResponse
// 400:
// 404:
// 409:
// 500:

// Created
// The operation was processed successfully
//
// A BookingSuccessResponse Object
// swagger:response BookingSuccessResponse
type BookingSuccessResponse struct {
	// in:body
	booking.BookingResponse
}

// swagger:parameters Booking
type BookingRequest struct {
	// A  booking object
	// in:body
	// required:true
	Body booking.BookingRequest
}
