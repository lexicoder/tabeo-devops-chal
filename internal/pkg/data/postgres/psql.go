package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"spacetrouble/internal/pkg/entity"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	ans := Store{
		db: db,
	}
	return &ans
}

func (o *Store) AllBookingsPaginated(ctx context.Context, afterTime time.Time, afterUuid string, limit int) ([]entity.Booking, error) {
	q := `SELECT 
			B.id, B.status, B.created_at,
			U.id, U.first_name, U.last_name, U.gender, U.birthday,
			F.id, F.launchpad_id, F.launch_date,
			D.id, D.name
		FROM bookings B 
		JOIN users U ON U.id = B.user_id
		JOIN flights F ON F.id = B.flight_id
		JOIN destinations D ON D.id = F.destination_id
		`
	var args []interface{}
	if !afterTime.IsZero() && afterUuid != "" {
		q += " WHERE B.created_at > $1 AND B.id > $2"
		args = append(args, afterTime, afterUuid)
	}

	q += " ORDER BY B.created_at, B.id"
	q += fmt.Sprintf(" LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := o.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.Booking
	for rows.Next() {
		var item entity.Booking
		err := rows.Scan(
			&item.ID, &item.Status, &item.CreatedAt,
			&item.User.ID, &item.User.FirstName, &item.User.LastName,
			&item.User.Gender, &item.User.Birthday,
			&item.Flight.ID, &item.Flight.LaunchpadID, &item.Flight.Date,
			&item.Flight.Destination.ID, &item.Flight.Destination.Name,
		)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (o *Store) GetLaunchPadWeekAvailability(ctx context.Context, launchpadId, destinationId string,
	t time.Time) (bool, error) {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	ans, err := o.getLaunchPadWeekAvailabiltyTx(ctx, tx, launchpadId, destinationId, t)
	if err != nil {
		return ans, err
	}
	return ans, tx.Commit(ctx)
}

func (o *Store) getLaunchPadWeekAvailabiltyTx(ctx context.Context, tx pgx.Tx,
	launchpadId, destinationId string, t time.Time) (bool, error) {
	var ans bool
	err := tx.QueryRow(ctx, `SELECT launch_in_same_week($1, $2, $3)`, launchpadId, destinationId, t).Scan(&ans)

	return ans, err
}

func (o *Store) GetAllDestinations(ctx context.Context) ([]entity.Destination, error) {
	q := `SELECT id, name FROM destinations`
	rows, err := o.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.Destination
	for rows.Next() {
		var item entity.Destination
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (o *Store) CreateDestination(ctx context.Context, name string) (entity.Destination, error) {
	q := `INSERT INTO destinations VALUES($1, $2)`
	dst := entity.Destination{
		ID:   uuid.New(),
		Name: name,
	}
	_, err := o.db.Exec(ctx, q, dst.ID, dst.Name)
	return dst, err
}

func (o *Store) GetDestinationById(ctx context.Context, id string) (entity.Destination, error) {
	q := `SELECT id, name FROM destinations WHERE id = $1`
	var dest entity.Destination
	if err := o.db.QueryRow(ctx, q, id).Scan(&dest.ID, &dest.Name); err != nil {
		return dest, err
	}
	return dest, nil

}

func (o *Store) CreateBooking(ctx context.Context, u entity.User, f entity.Flight) (entity.Booking, error) {
	uq := `INSERT INTO users(id, first_name, last_name, gender, birthday)
			VALUES($1, $2, $3, $4, $5) ON CONFLICT(id) DO NOTHING`
	fq := `INSERT INTO flights(id, launchpad_id, destination_id, launch_date)
VALUES($1, $2, $3, $4)`
	bq := `INSERT INTO bookings(id, user_id, flight_id, status, created_at) VALUES($1, $2, $3, $4, $5)`

	// TODO maybe check for business rules violations within the transaction
	// Now will return a constraint violation error.
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return entity.Booking{}, err
	}
	nb := entity.Booking{
		ID:        uuid.New(),
		User:      u,
		Flight:    f,
		Status:    entity.BookingStatusActive,
		CreatedAt: time.Now().UTC(),
	}
	defer tx.Rollback(ctx)
	if f.IsIDEmpty() {
		f.ID = uuid.New()
		nb.Flight.ID = f.ID
		if _, err := tx.Exec(ctx, fq, f.ID, f.LaunchpadID, f.Destination.ID, f.Date); err != nil {
			return nb, err
		}
	}
	if _, err := tx.Exec(ctx, uq, u.ID, u.FirstName, u.LastName, u.Gender, u.Birthday); err != nil {
		return nb, err
	}
	if _, err := tx.Exec(ctx, bq, nb.ID, nb.User.ID, nb.Flight.ID, nb.Status, nb.CreatedAt); err != nil {
		return nb, err
	}
	return nb, tx.Commit(ctx)
}

func (o *Store) SelectFlights(ctx context.Context, filters map[string]interface{}) ([]entity.Flight, error) {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	flights, err := o.selectFlightsTx(ctx, tx, filters)
	if err != nil {
		return nil, err
	}
	return flights, tx.Commit(ctx)
}

func (o *Store) buildSelectFlightQ(filters map[string]interface{}) (string, []interface{}) {
	q := `SELECT 
			F.id, F.launchpad_id, F.launch_date,
			D.id as destination_id, D.name as destination_name
			FROM flights F
			JOIN destinations D ON D.id = F.destination_id`
	bookingStatus, hasBookingStatus := filters["bookings.status"]
	if hasBookingStatus {
		q += ` JOIN bookings B ON B.flight_id = F.id`
	}
	var whereConds []string
	var args []interface{}
	for k, v := range filters {
		if !strings.HasPrefix(k, "bookings.") {
			whereConds = append(whereConds, fmt.Sprintf("F.%s=$%d", k, len(args)+1))
			args = append(args, v)
		}
	}
	if hasBookingStatus {
		whereConds = append(whereConds, fmt.Sprintf("B.status=$%d", len(args)+1))
		args = append(args, bookingStatus)
	}
	if len(whereConds) > 0 {
		q += " WHERE " + strings.Join(whereConds, " AND ")
	}
	if hasBookingStatus {
		q += " GROUP BY F.id, D.id"
	}
	return q, args
}

func (o *Store) selectFlightsTx(ctx context.Context, tx pgx.Tx, filters map[string]interface{}) ([]entity.Flight, error) {
	q, args := o.buildSelectFlightQ(filters)
	rows, err := tx.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.Flight
	for rows.Next() {
		var flight entity.Flight
		err := rows.Scan(
			&flight.ID, &flight.LaunchpadID, &flight.Date,
			&flight.Destination.ID, &flight.Destination.Name,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, flight)
	}
	return items, rows.Err()
}
