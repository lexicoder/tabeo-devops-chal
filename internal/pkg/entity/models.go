package entity

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	BookingStatusActive = "active"
)

type Store interface {
	CreateDestination(ctx context.Context, name string) (Destination, error)
	GetAllDestinations(ctx context.Context) ([]Destination, error)
	GetDestinationById(ctx context.Context, id string) (Destination, error)
	CreateBooking(ctx context.Context, u User, f Flight) (Booking, error)
	SelectFlights(ctx context.Context, filters map[string]interface{}) ([]Flight, error)
	GetLaunchPadWeekAvailability(ctx context.Context, launchpadId, destinationId string, t time.Time) (bool, error)
	AllBookingsPaginated(ctx context.Context, afterTime time.Time, afterUuid string, limit int) ([]Booking, error)
}

type Destination struct {
	ID   uuid.UUID
	Name string
}

type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Gender    string
	Birthday  time.Time
}

func (o User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		Birthday string
		Gender   string
		Alias
	}{
		Birthday: o.Birthday.Format("2006-01-02"),
		Gender:   genderFull(o.Gender),
		Alias:    (Alias)(o),
	})
}

func genderFull(v string) string {
	switch v {
	case "m":
		return "male"
	case "f":
		return "female"
	case "o":
		return "other"
	default:
		return ""
	}
}

type Flight struct {
	ID          uuid.UUID
	LaunchpadID string
	Destination Destination
	Date        time.Time
}

func (o Flight) MarshalJSON() ([]byte, error) {
	type Alias Flight
	return json.Marshal(&struct {
		Date string
		Alias
	}{
		Date:  o.Date.Format("2006-01-02"),
		Alias: (Alias)(o),
	})
}

func (o *Flight) IsIDEmpty() bool {
	return o.ID == uuid.Nil
}

type Booking struct {
	ID        uuid.UUID
	User      User
	Flight    Flight
	Status    string
	CreatedAt time.Time
}
