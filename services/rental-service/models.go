package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ── Cross-service types ────────────────────────────────────────────────────

// Item is fetched from item-service via HTTP, not from a local table.
type Item struct {
	ID             string      `json:"id"`
	PickupPoint    PickupPoint `json:"pickup_point"`
	PricePerHour   int         `json:"price_per_hour"`
	PricePerDay    int         `json:"price_per_day"`
	IsPlatformItem bool        `json:"is_platform_item"`
	Title          string      `json:"title"`
	Category       string      `json:"category"`
}

type PickupPoint struct {
	ID      string `json:"id"`
	City    string `json:"city"`
	Address string `json:"address"`
}

// ── Local DB types ─────────────────────────────────────────────────────────

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

type MyRental struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	StartAt    time.Time `json:"start_at"`
	EndAt      time.Time `json:"end_at"`
	TotalPrice int       `json:"total_price"`
	Item       struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Category string `json:"category"`
	} `json:"item"`
	PickupPoint struct {
		City    string `json:"city"`
		Address string `json:"address"`
	} `json:"pickup_point"`
}

// RentalReminder now carries only the IDs; titles/emails are fetched via HTTP.
type RentalReminder struct {
	RentalID string
	ItemID   string
	RenterID string
	EndAt    time.Time
}

// ── HTTP calls to peer services ────────────────────────────────────────────

func (app *application) fetchItem(itemID string) (*Item, error) {
	url := fmt.Sprintf("%s/items/%s", app.cfg.itemServiceURL, itemID)
	resp, err := app.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("item-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("item not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("item-service returned %d", resp.StatusCode)
	}
	var item Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (app *application) fetchUserEmail(userID string) (string, error) {
	url := fmt.Sprintf("%s/internal/users/%s", app.cfg.authServiceURL, userID)
	resp, err := app.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("auth-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth-service returned %d", resp.StatusCode)
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.Email, nil
}

// ── Local DB queries ───────────────────────────────────────────────────────

func (app *application) dbCreateRental(rental *Rental) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return app.db.QueryRowContext(ctx,
		`INSERT INTO rentals(item_id,renter_id,pickup_point_id,start_at,end_at,total_price,platform_fee,owner_income)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,status`,
		rental.ItemID, rental.RenterID, rental.PickupPointID, rental.StartAt, rental.EndAt,
		rental.TotalPrice, rental.PlatformFee, rental.OwnerIncome,
	).Scan(&rental.ID, &rental.Status)
}

func (app *application) dbGetRentalsByUser(userID string) ([]*MyRental, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		`SELECT id,status,start_at,end_at,total_price,item_id,renter_id,pickup_point_id
		 FROM rentals WHERE renter_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rentals := []*MyRental{}
	for rows.Next() {
		var ren MyRental
		var itemID, renterID, pickupPointID string
		if err := rows.Scan(
			&ren.ID, &ren.Status, &ren.StartAt, &ren.EndAt, &ren.TotalPrice,
			&itemID, &renterID, &pickupPointID,
		); err != nil {
			return nil, err
		}
		ren.Item.ID = itemID
		rentals = append(rentals, &ren)
	}
	return rentals, rows.Err()
}

func (app *application) dbGetRental(id string) (*Rental, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var r Rental
	err := app.db.QueryRowContext(ctx,
		`SELECT id,item_id,renter_id,pickup_point_id,start_at,end_at,total_price,platform_fee,owner_income,status
		 FROM rentals WHERE id=$1`, id,
	).Scan(&r.ID, &r.ItemID, &r.RenterID, &r.PickupPointID, &r.StartAt, &r.EndAt,
		&r.TotalPrice, &r.PlatformFee, &r.OwnerIncome, &r.Status)
	return &r, err
}

func (app *application) dbUpdateRentalStatus(userID, rentalID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := app.db.ExecContext(ctx,
		`UPDATE rentals SET status=$1 WHERE renter_id=$2 AND id=$3`, status, userID, rentalID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.New("rental not found")
	}
	return nil
}

func (app *application) dbGetRentalsForReminder() ([]*RentalReminder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		`SELECT id, item_id, renter_id, end_at
		 FROM rentals
		 WHERE status='ACTIVE'
		   AND reminder_sent=FALSE
		   AND end_at <= now() + INTERVAL '30 minutes'
		   AND end_at > now()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reminders := []*RentalReminder{}
	for rows.Next() {
		var rem RentalReminder
		if err := rows.Scan(&rem.RentalID, &rem.ItemID, &rem.RenterID, &rem.EndAt); err != nil {
			return nil, err
		}
		reminders = append(reminders, &rem)
	}
	return reminders, rows.Err()
}

func (app *application) dbMarkReminderSent(rentalID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := app.db.ExecContext(ctx, `UPDATE rentals SET reminder_sent=TRUE WHERE id=$1`, rentalID)
	return err
}

func (app *application) dbPing() error {
	return app.db.Ping()
}

