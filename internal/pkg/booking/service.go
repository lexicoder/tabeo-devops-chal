package booking

import (
	"context"
	"time"

	"github.com/google/uuid"

	"spacetrouble/internal/pkg/entity"
)

type BookingService interface {
	MakeBooking(ctx context.Context, req BookingRequest) (BookingResponse, error)
	AllBookings(ctx context.Context, req GetBookingsReq) (AllBookingsResponse, error)
}

type SpaceX interface {
	IsLaunchpadAvailable(ctx context.Context, launchpadID string, ts time.Time) (bool, error)
}

type bookingSrv struct {
	store  entity.Store
	spacex SpaceX
}

func NewBookingService(store entity.Store, spacex SpaceX) *bookingSrv {
	ans := bookingSrv{
		store:  store,
		spacex: spacex,
	}
	return &ans
}

func (o *bookingSrv) AllBookings(ctx context.Context, req GetBookingsReq) (AllBookingsResponse, error) {
	ans := AllBookingsResponse{
		Bookings: make([]BookingResponse, 0),
		Limit:    req.Limit,
	}
	bookings, err := o.store.AllBookingsPaginated(ctx, req.Ts, req.Uuid, req.Limit)
	if err != nil {
		return ans, err
	}
	for i := range bookings {
		ans.Bookings = append(ans.Bookings, BookingResponse{Booking: bookings[i]})
	}
	if len(ans.Bookings) > 0 {
		ans.Cursor = encodeCursor(
			ans.Bookings[len(ans.Bookings)-1].CreatedAt,
			ans.Bookings[len(ans.Bookings)-1].ID.String(),
		)
	}
	return ans, nil
}

func (o *bookingSrv) MakeBooking(ctx context.Context, req BookingRequest) (BookingResponse, error) {
	var ans BookingResponse
	destination, err := o.store.GetDestinationById(ctx, req.DestinationID)
	if err != nil {
		return ans, ErrMissingDestination
	}

	if err := o.islaunchpadUsed(ctx, req.LaunchpadID, destination.ID.String(), req.LaunchDate.Time); err != nil {
		return ans, err
	}

	flight, err := o.currentFlightLaunchPad(ctx, req.LaunchpadID, destination.ID.String(), req.LaunchDate.Time)
	if err != nil {
		return ans, err
	}

	// When we don't already have a flight.ID then we check the availability of spaceX
	if flight.IsIDEmpty() {

		// before that we check that we can make a booking for the destination for this week.
		// if there is already on from the same launchpad abort
		err = o.sameDestinationLaunchPad(ctx, req.LaunchpadID, destination.ID.String(), req.LaunchDate.Time)
		if err != nil {
			return ans, err
		}
		flight, err = o.createFlightSpaceX(ctx, req.LaunchpadID, destination, req.LaunchDate.Time)
		if err != nil {
			return ans, err
		}
	}

	user := entity.User{
		ID:        uuid.New(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Gender:    string(req.Gender[0]),
		Birthday:  req.Birthday.Time,
	}
	// we can now create the booking
	newBooking, err := o.store.CreateBooking(ctx, user, flight)
	if err != nil {
		return ans, err
	}
	ans.Booking = newBooking
	return ans, nil
}

// we forbid booking from a launchpad that it's already used
func (o *bookingSrv) islaunchpadUsed(ctx context.Context, launchpadId string, destinationId string, date time.Time) error {
	flights, err := o.store.SelectFlights(
		ctx,
		map[string]interface{}{
			"launchpad_id": launchpadId,
			"launch_date":  date,
		},
	)
	if err != nil {
		return err
	}
	if len(flights) > 0 && flights[0].Destination.ID.String() != destinationId {
		return ErrLaunchPadUnavailable
	}

	return nil
}

// We search in our database if we have a flight with active booking
// for the launchpad destination and launch date.
func (o *bookingSrv) currentFlightLaunchPad(ctx context.Context, launchpadId, destinationId string, date time.Time) (flight entity.Flight, err error) {
	var flights []entity.Flight
	flights, err = o.store.SelectFlights(
		ctx,
		map[string]interface{}{
			"launchpad_id":    launchpadId,
			"destination_id":  destinationId,
			"launch_date":     date,
			"bookings.status": entity.BookingStatusActive,
		},
	)
	if err != nil {
		return
	}
	if len(flights) == 1 {
		flight = flights[0]
	}
	return
}

func (o *bookingSrv) sameDestinationLaunchPad(ctx context.Context, launchpadId, destinationId string, date time.Time) error {
	ok, err := o.store.GetLaunchPadWeekAvailability(ctx, launchpadId, destinationId, date)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLaunchPadUnavailable
	}
	return nil
}

func (o *bookingSrv) createFlightSpaceX(ctx context.Context, launchpadId string, dst entity.Destination,
	date time.Time) (flight entity.Flight, err error) {
	var isAvailable bool
	isAvailable, err = o.spacex.IsLaunchpadAvailable(ctx, launchpadId, date)
	if err != nil {
		return
	}
	if !isAvailable {
		err = ErrLaunchPadUnavailable
		return
	}
	flight = entity.Flight{
		LaunchpadID: launchpadId,
		Destination: dst,
		Date:        date,
	}
	return
}
