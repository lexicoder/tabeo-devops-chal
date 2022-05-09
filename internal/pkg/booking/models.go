package booking

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"spacetrouble/internal/pkg/entity"
)

const (
	dateLayoutFmt = "2006-01-02"
)

type Date struct {
	time.Time
}

func (d Date) String() string {
	return d.Time.Format(dateLayoutFmt)
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	var err error
	d.Time, err = time.Parse(dateLayoutFmt, s)
	if err != nil {
		return err
	}
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(d.Time.Format(dateLayoutFmt)), nil
}

type BookingRequest struct {
	ID            string `json:"id,omitempty"`
	FirstName     string
	LastName      string
	Gender        string
	Birthday      Date
	LaunchpadID   string
	DestinationID string
	LaunchDate    Date `json:"Date"`
}

var lock *sync.RWMutex = &sync.RWMutex{}
var supportedGenders map[string]bool = map[string]bool{
	"female": true,
	"male":   true,
	"other":  true,
}

func (o *BookingRequest) Validate() error {
	if len(o.FirstName) == 0 || len(o.FirstName) > 50 {
		return errors.New("FirstName must be more than 0 and less than 50 chars")
	}
	if len(o.LastName) == 0 || len(o.LastName) > 50 {
		return errors.New("LastName must be more than 0 and less than 50 chars")
	}
	if o.LaunchDate.IsZero() {
		return errors.New("Date is empty")
	}
	if o.LaunchDate.Before(time.Now()) {
		return errors.New("Date is in the past")
	}
	if o.Birthday.IsZero() {
		return errors.New("empty Birthday")
	}

	if o.Birthday.After(time.Now()) {
		return errors.New("Birthday is in the past")
	}

	lock.RLock()
	ok := supportedGenders[o.Gender]
	lock.RUnlock()
	if !ok {
		return errors.New("invalid Gender")
	}
	if len(o.LaunchpadID) != 24 {
		return errors.New("launchPadID must have length 24")
	}
	if _, err := uuid.Parse(o.DestinationID); err != nil {
		return errors.New("invalid uuid for DestinationID")
	}

	return nil
}

type BookingResponse struct {
	entity.Booking
}

type AllBookingsResponse struct {
	Bookings []BookingResponse `json:"bookings"`
	Limit    int               `json:"limit"`
	Cursor   string            `json:"cursor"`
}

type GetBookingsReq struct {
	Limit int
	Uuid  string
	Ts    time.Time
}

func decodeCursor(encoded string) (ans time.Time, uuid string, err error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return
	}
	arr := strings.Split(string(b), ",")
	if len(arr) != 2 {
		err = errors.New("invalid cursor")
		return
	}
	ans, err = time.Parse(time.RFC3339Nano, arr[0])
	if err != nil {
		return
	}
	uuid = arr[1]
	return
}

func encodeCursor(t time.Time, uuid string) string {
	key := fmt.Sprintf("%s,%s", t.Format(time.RFC3339Nano), uuid)
	return base64.StdEncoding.EncodeToString([]byte(key))
}
