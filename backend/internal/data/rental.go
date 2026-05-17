package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Rental struct {
	ID            string    `json:"id"`
	ItemID        string    `json:"item_id"`
	RenterID      string    `json:"renter_id"`
	PickupPointID string    `json:"pickup_point_id"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	TotalPrice    int       `json:"total_price"`
	PlatformFee   int       `json:"platform_fee"`
	OwnerIncome   int       `json:"owner_income"`
	Status        string    `json:"status"`
}

type MyRentalResponse struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	StartAt    time.Time `json:"start_at"`
	EndAt      time.Time `json:"end_at"`
	TotalPrice int       `json:"total_price"`

	Item struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Category string `json:"category"`
	} `json:"item"`

	PickupPoint struct {
		City    string `json:"city"`
		Address string `json:"address"`
	} `json:"pickup_point"`
}

type RentalReminder struct {
	RentalID  string    `json:"rental_id"`
	ItemTitle string    `json:"item_title"`
	Email     string    `json:"email"`
	EndAt     time.Time `json:"end_at"`
}
type RentalModel struct {
	DB *sql.DB
}

func (r *RentalModel) Create(rental *Rental) error {
	query := `INSERT INTO rentals (item_id, renter_id, pickup_point_id, start_at, end_at, total_price, platform_fee, owner_income) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,status`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{rental.ItemID, rental.RenterID, rental.PickupPointID, rental.StartAt, rental.EndAt, rental.TotalPrice, rental.PlatformFee, rental.OwnerIncome}

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&rental.ID, &rental.Status)
	if err != nil {
		return err
	}
	return nil
}

func (r *RentalModel) GetListByUserID(userID string) ([]*MyRentalResponse, error) {
	query := `SELECT rentals.id,rentals.status,rentals.start_at,rentals.end_at,rentals.total_price,items.id,items.title,items.category,pickup_points.city,pickup_points.address 
			FROM rentals 
			JOIN items ON rentals.item_id = items.id 
			JOIN pickup_points ON rentals.pickup_point_id = pickup_points.id 
			WHERE rentals.renter_id = $1 
			
			`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, userID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	rentals := []*MyRentalResponse{}
	for rows.Next() {
		var rental MyRentalResponse

		err := rows.Scan(
			&rental.ID,
			&rental.Status,
			&rental.StartAt,
			&rental.EndAt,
			&rental.TotalPrice,
			&rental.Item.ID,
			&rental.Item.Title,
			&rental.Item.Category,
			&rental.PickupPoint.City,
			&rental.PickupPoint.Address,
		)
		if err != nil {
			return nil, err
		}

		rentals = append(rentals, &rental)

	}

	return rentals, nil
}

func (r *RentalModel) UpdateStatus(userID, rentalID, status string) error {
	query := `UPDATE rentals SET status = $1 WHERE renter_id = $2 AND id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{status, userID, rentalID}

	result, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Not Found")
	}

	return nil
}

func (r *RentalModel) GetByID(rentalID string) (*Rental, error) {
	query := `SELECT id, item_id, renter_id, pickup_point_id, start_at, end_at, total_price, platform_fee, owner_income, status FROM rentals WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rental Rental

	err := r.DB.QueryRowContext(ctx, query, rentalID).Scan(
		&rental.ID,
		&rental.ItemID,
		&rental.RenterID,
		&rental.PickupPointID,
		&rental.StartAt,
		&rental.EndAt,
		&rental.TotalPrice,
		&rental.PlatformFee,
		&rental.Status,
	)
	if err != nil {
		return nil, err
	}

	return &rental, nil
}

func (r *RentalModel) GetRentalsForReminder() ([]*RentalReminder, error) {
	query := `SELECT rentals.id,items.title,users.email,rentals.end_at FROM users JOIN rentals ON users.id = rentals.renter_id JOIN items ON items.id = rentals.item_id  WHERE rentals.status IN ('ACTIVE') AND rentals.reminder_sent = FALSE AND rentals.end_at <= now() + INTERVAL '30 minutes' AND rentals.end_at > now()`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rentals := []*RentalReminder{}

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rental RentalReminder

		err := rows.Scan(
			&rental.RentalID,
			&rental.ItemTitle,
			&rental.Email,
			&rental.EndAt,
		)

		if err != nil {
			return nil, err
		}

		rentals = append(rentals, &rental)
	}

	return rentals, nil
}

func (r *RentalModel) UpdateReminderSent(rentalID string) error {
	query := `UPDATE rentals SET reminder_sent = TRUE WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.DB.ExecContext(ctx, query, rentalID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("rental not found")
	}
	return nil

}
