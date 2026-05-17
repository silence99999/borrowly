package data

import (
	"context"
	"database/sql"
	"time"
)

type PickupPoint struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type PickupPointModel struct {
	DB *sql.DB
}

func (pp *PickupPointModel) GetAllActive() ([]*PickupPoint, error) {
	query := `SELECT id, address, city, is_active, created_at FROM pickup_points WHERE is_active = TRUE ORDER BY city, address`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := pp.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pickupPoints := []*PickupPoint{}
	for rows.Next() {
		var pickupPoint PickupPoint

		err := rows.Scan(
			&pickupPoint.ID,
			&pickupPoint.Address,
			&pickupPoint.City,
			&pickupPoint.IsActive,
			&pickupPoint.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		pickupPoints = append(pickupPoints, &pickupPoint)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pickupPoints, nil
}

func (pp *PickupPointModel) Create(pickupPoint *PickupPoint) error {
	query := `INSERT INTO pickup_points (address, city,is_active) VALUES ($1,$2,$3) RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{pickupPoint.Address, pickupPoint.City, pickupPoint.IsActive}

	err := pp.DB.QueryRowContext(ctx, query, args...).Scan(&pickupPoint.ID)
	if err != nil {
		return err
	}
	return nil
}
