package booking

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"spacetrouble/internal/pkg/entity"

	"spacetrouble/internal/pkg/data/postgres"

	"spacetrouble/pkg/testutils"
)

type SpaceXMockAvailable struct{}

func (o *SpaceXMockAvailable) IsLaunchpadAvailable(ctx context.Context, launchpadId string, date time.Time) (bool, error) {
	return true, nil
}

type SpaceXMockUnAvailable struct{}

func (o *SpaceXMockUnAvailable) IsLaunchpadAvailable(ctx context.Context, launchpadId string, date time.Time) (bool, error) {
	return false, nil
}

type SpaceXMockError struct{}

func (o *SpaceXMockError) IsLaunchpadAvailable(ctx context.Context, launchpadId string, date time.Time) (bool, error) {
	return false, errors.New("spaceX api error")
}

func genLaunchId() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)[:24]
}

func TestMain(m *testing.M) {
	workingDir, _ := os.Getwd()
	rootDir := strings.Replace(workingDir, "internal/pkg/booking", "", 1)

	ctx := context.Background()
	postgresContainer := testutils.SpinPostgresContainer(ctx, rootDir)

	defer postgresContainer.Terminate(ctx)

	exitCode := m.Run()

	os.Exit(exitCode)
}

func getStoreAndDb() (entity.Store, *pgxpool.Pool, error) {
	db, err := testutils.GetTestDb()
	if err != nil {
		return nil, nil, err
	}
	if err := db.Ping(context.Background()); err != nil {
		return nil, nil, err
	}
	store := postgres.NewStore(db)
	return store, db, nil
}

func createDestinations(store entity.Store) ([]entity.Destination, error) {
	return store.GetAllDestinations(context.Background())
}

func cleanDatabase(db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(),
		"truncate users cascade;truncate flights cascade; truncate bookings;",
	)
	if err != nil {
		panic(err)
	}
}

func checkBookingCount(db *pgxpool.Pool, expected int) (int, bool, error) {
	var cnt int
	if err := db.QueryRow(context.Background(), `select count(1) from bookings`).Scan(&cnt); err != nil {
		return 0, false, err
	}
	return cnt, cnt == expected, nil
}

func TestMakeBookingNewAndAvailable(t *testing.T) {
	store, db, err := getStoreAndDb()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	defer cleanDatabase(db)

	availableDestinations, err := createDestinations(store)
	if err != nil {
		t.Error(err)
		return
	}

	spaceClient := &SpaceXMockAvailable{}
	srv := NewBookingService(store, spaceClient)

	birthday, _ := time.Parse(dateLayoutFmt, "13/11/1923")
	launchDate, _ := time.Parse(dateLayoutFmt, "06/04/2021")
	req := BookingRequest{
		FirstName:     "Giorgos",
		LastName:      "Papadopoulos",
		Gender:        "male",
		Birthday:      Date{Time: birthday},
		LaunchpadID:   genLaunchId(),
		DestinationID: availableDestinations[0].ID.String(),
		LaunchDate:    Date{Time: launchDate},
	}

	newBooking, err := srv.MakeBooking(context.Background(), req)
	if err != nil {
		t.Error(err)
		return
	}

	if newBooking.User.FirstName != "Giorgos" && newBooking.User.LastName != "Papadopoulos" {
		t.Errorf("first names differ")
		return
	}

	numBookings, ok, err := checkBookingCount(db, 1)
	if err != nil {
		t.Error(err)
		return
	}

	if !ok {
		t.Errorf("expected to insert 1 flight found %d", numBookings)
		return
	}
}

func TestMakeBookingMissingDestination(t *testing.T) {
	store, db, err := getStoreAndDb()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	defer cleanDatabase(db)

	_, err = createDestinations(store)
	if err != nil {
		t.Error(err)
		return
	}

	spaceClient := &SpaceXMockAvailable{}
	srv := NewBookingService(store, spaceClient)

	birthday, _ := time.Parse(dateLayoutFmt, "13/11/1923")
	launchDate, _ := time.Parse(dateLayoutFmt, "06/04/2021")
	req := BookingRequest{
		FirstName:     "Giorgos",
		LastName:      "Papadopoulos",
		Gender:        "male",
		Birthday:      Date{Time: birthday},
		LaunchpadID:   genLaunchId(),
		DestinationID: uuid.New().String(),
		LaunchDate:    Date{Time: launchDate},
	}

	_, err = srv.MakeBooking(context.Background(), req)
	if err == nil {
		t.Errorf("expected error but got nil")
		return
	}
}

func TestMakeBookingWhenThereAreAlreadyBookingsForFlight(t *testing.T) {
	store, db, err := getStoreAndDb()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	defer cleanDatabase(db)

	availableDestinations, err := createDestinations(store)
	if err != nil {
		t.Error(err)
		return
	}

	spaceClient := &SpaceXMockAvailable{}
	srv := NewBookingService(store, spaceClient)

	birthday, _ := time.Parse(dateLayoutFmt, "13/11/1923")
	launchDate, _ := time.Parse(dateLayoutFmt, "06/04/2021")
	req := BookingRequest{
		FirstName:     "Giorgos",
		LastName:      "Papadopoulos",
		Gender:        "male",
		Birthday:      Date{Time: birthday},
		LaunchpadID:   genLaunchId(),
		DestinationID: availableDestinations[0].ID.String(),
		LaunchDate:    Date{Time: launchDate},
	}

	_, err = srv.MakeBooking(context.Background(), req)
	if err != nil {
		t.Error(err)
		return
	}

	req2 := req
	req2.FirstName = "John"

	_, err = srv.MakeBooking(context.Background(), req2)
	if err != nil {
		t.Error(err)
		return
	}

	cnt, ok, err := checkBookingCount(db, 2)
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Errorf("expected 2 bookings but got %d", cnt)
		return
	}
}

func TestMakeBookingWhenLaunchpadHasOtherBookingSameDate(t *testing.T) {
	store, db, err := getStoreAndDb()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	defer cleanDatabase(db)
	availableDestinations, err := createDestinations(store)
	if err != nil {
		t.Error(err)
		return
	}

	spaceClient := &SpaceXMockAvailable{}
	srv := NewBookingService(store, spaceClient)

	birthday, _ := time.Parse(dateLayoutFmt, "13/11/1923")
	launchDate, _ := time.Parse(dateLayoutFmt, "06/04/2021")
	req := BookingRequest{
		FirstName:     "Giorgos",
		LastName:      "Papadopoulos",
		Gender:        "male",
		Birthday:      Date{Time: birthday},
		LaunchpadID:   genLaunchId(),
		DestinationID: availableDestinations[0].ID.String(),
		LaunchDate:    Date{Time: launchDate},
	}

	_, err = srv.MakeBooking(context.Background(), req)
	if err != nil {
		t.Error(err)
		return
	}

	cnt, ok, err := checkBookingCount(db, 1)
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Errorf("expected 1 bookings but got %d", cnt)
		return
	}

	req2 := req
	req2.DestinationID = availableDestinations[1].ID.String()
	_, err = srv.MakeBooking(context.Background(), req2)
	if err == nil {
		t.Errorf("expected not to be able to make a booking to another dest using this launchpad")
		return
	}

	cnt, ok, err = checkBookingCount(db, 1)
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Errorf("expected 1 bookings but got %d", cnt)
		return
	}
}

func TestMakeBookingFromSameLaunchpadToSameDestinationInSameWeek(t *testing.T) {
	store, db, err := getStoreAndDb()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	defer cleanDatabase(db)
	availableDestinations, err := createDestinations(store)
	if err != nil {
		t.Error(err)
		return
	}

	spaceClient := &SpaceXMockAvailable{}
	srv := NewBookingService(store, spaceClient)

	birthday, _ := time.Parse(dateLayoutFmt, "13/11/1923")
	launchDate, _ := time.Parse(dateLayoutFmt, "06/04/2021")
	req := BookingRequest{
		FirstName:     "Giorgos",
		LastName:      "Papadopoulos",
		Gender:        "male",
		Birthday:      Date{Time: birthday},
		LaunchpadID:   genLaunchId(),
		DestinationID: availableDestinations[0].ID.String(),
		LaunchDate:    Date{Time: launchDate},
	}

	_, err = srv.MakeBooking(context.Background(), req)
	if err != nil {
		t.Error(err)
		return
	}

	req2 := req
	req2.LaunchDate = Date{Time: launchDate.Add(48 * time.Hour)}

	_, err = srv.MakeBooking(context.Background(), req2)

	if err == nil {
		t.Error(err)
		return
	}

	cnt, ok, err := checkBookingCount(db, 1)
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Errorf("expected 1 bookings but got %d", cnt)
		return
	}
}

func TestMakeBookingFromSameLaunchpadSameDestinationNextWeek(t *testing.T) {
	store, db, err := getStoreAndDb()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	defer cleanDatabase(db)
	availableDestinations, err := createDestinations(store)
	if err != nil {
		t.Error(err)
		return
	}

	spaceClient := &SpaceXMockAvailable{}
	srv := NewBookingService(store, spaceClient)

	birthday, _ := time.Parse(dateLayoutFmt, "13/11/1923")
	launchDate, _ := time.Parse(dateLayoutFmt, "06/04/2021")
	req := BookingRequest{
		FirstName:     "Giorgos",
		LastName:      "Papadopoulos",
		Gender:        "male",
		Birthday:      Date{Time: birthday},
		LaunchpadID:   genLaunchId(),
		DestinationID: availableDestinations[0].ID.String(),
		LaunchDate:    Date{Time: launchDate},
	}

	_, err = srv.MakeBooking(context.Background(), req)
	if err != nil {
		t.Error(err)
		return
	}

	req2 := req
	req2.LaunchDate = Date{Time: launchDate.Add(168 * time.Hour)}

	_, err = srv.MakeBooking(context.Background(), req2)

	if err != nil {
		t.Error(err)
		return
	}

	cnt, ok, err := checkBookingCount(db, 2)
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Errorf("expected 1 bookings but got %d", cnt)
		return
	}
}

func TestMakeBookingWithNoSpaceXAvailability(t *testing.T) {
	store, db, err := getStoreAndDb()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	defer cleanDatabase(db)
	availableDestinations, err := createDestinations(store)
	if err != nil {
		t.Error(err)
		return
	}

	spaceClient := &SpaceXMockUnAvailable{}
	srv := NewBookingService(store, spaceClient)

	birthday, _ := time.Parse(dateLayoutFmt, "13/11/1923")
	launchDate, _ := time.Parse(dateLayoutFmt, "06/04/2021")
	req := BookingRequest{
		FirstName:     "Giorgos",
		LastName:      "Papadopoulos",
		Gender:        "male",
		Birthday:      Date{Time: birthday},
		LaunchpadID:   genLaunchId(),
		DestinationID: availableDestinations[0].ID.String(),
		LaunchDate:    Date{Time: launchDate},
	}

	_, err = srv.MakeBooking(context.Background(), req)
	if err == nil {
		t.Error("expected not to save because spacex unavailable")
		return
	}
}

func TestMakeBoookingWithSpaceXReturnError(t *testing.T) {
	store, db, err := getStoreAndDb()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	defer cleanDatabase(db)
	availableDestinations, err := createDestinations(store)
	if err != nil {
		t.Error(err)
		return
	}

	spaceClient := &SpaceXMockError{}
	srv := NewBookingService(store, spaceClient)

	birthday, _ := time.Parse(dateLayoutFmt, "13/11/1923")
	launchDate, _ := time.Parse(dateLayoutFmt, "06/04/2021")
	req := BookingRequest{
		FirstName:     "Giorgos",
		LastName:      "Papadopoulos",
		Gender:        "male",
		Birthday:      Date{Time: birthday},
		LaunchpadID:   genLaunchId(),
		DestinationID: availableDestinations[0].ID.String(),
		LaunchDate:    Date{Time: launchDate},
	}

	_, err = srv.MakeBooking(context.Background(), req)
	if err == nil {
		t.Error("expected not to save because spacex returned error")
		return
	}
}
