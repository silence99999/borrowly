package data

import (
	"context"
	"database/sql"
	"time"
)

type Item struct {
	ID             string    `json:"id"`
	OwnerID        string    `json:"owner_id"`
	PickupPointID  string    `json:"pickup_point_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	PricePerHour   int       `json:"price_per_hour"`
	PricePerDay    int       `json:"price_per_day"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	IsPlatformItem bool      `json:"is_platform_item"`
}

type ItemUpdate struct {
	PickupPointID *string `json:"pickup_point_id"`
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	Category      *string `json:"category"`
	PricePerHour  *int    `json:"price_per_hour"`
	PricePerDay   *int    `json:"price_per_day"`
}

type ItemDetails struct {
	ID             string      `json:"id"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	PickupPoint    PickupPoint `json:"pickup_point"`
	Category       string      `json:"category"`
	PricePerHour   int         `json:"price_per_hour"`
	PricePerDay    int         `json:"price_per_day"`
	City           string      `json:"pickup_city"`
	Owner          User        `json:"owner"`
	IsPlatformItem bool        `json:"is_platform_item"`
}

type ItemList struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Category       string `json:"category"`
	PricePerHour   int    `json:"price_per_hour"`
	PricePerDay    int    `json:"price_per_day"`
	City           string `json:"pickup_city"`
	IsPlatformItem bool   `json:"is_platform_item"`
}

type ItemModel struct {
	DB *sql.DB
}

func (im *ItemModel) Create(item *Item) error {
	query := `INSERT INTO items (owner_id, pickup_point_id, title, description, category, price_per_hour, price_per_day,is_platform_item)  VALUES ($1,$2,$3,$4,$5,$6,$7,$8) returning status,id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{item.OwnerID, item.PickupPointID, item.Title, item.Description, item.Category, item.PricePerHour, item.PricePerDay, item.IsPlatformItem}

	err := im.DB.QueryRowContext(ctx, query, args...).Scan(&item.Status, &item.ID)
	if err != nil {
		return err
	}

	return nil

}

func (im *ItemModel) GetAll() ([]*ItemList, error) {
	query := `SELECT items.id,items.title,items.category,items.price_per_hour,items.price_per_day,items.is_platform_item,pickup_points.city FROM items JOIN pickup_points ON items.pickup_point_id = pickup_points.id `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := im.DB.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []*ItemList{}

	for rows.Next() {
		var item ItemList
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Category,
			&item.PricePerHour,
			&item.PricePerDay,
			&item.IsPlatformItem,
			&item.City,
		)

		if err != nil {
			return nil, err
		}

		items = append(items, &item)

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil

}

func (im *ItemModel) GetByID(itemID string) (*ItemDetails, error) {
	query := `
		SELECT
			items.id,
			items.title,
			items.description,
			items.category,
			items.price_per_hour,
			items.price_per_day,
			items.is_platform_item,
			pickup_points.id,
			pickup_points.address,
			pickup_points.city,
			pickup_points.is_active,
			users.id,
			users.email
		FROM items
		JOIN pickup_points ON items.pickup_point_id = pickup_points.id
		JOIN users ON items.owner_id = users.id
		WHERE items.id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var itemDetails ItemDetails

	err := im.DB.QueryRowContext(ctx, query, itemID).Scan(
		&itemDetails.ID,
		&itemDetails.Title,
		&itemDetails.Description,
		&itemDetails.Category,
		&itemDetails.PricePerHour,
		&itemDetails.PricePerDay,
		&itemDetails.IsPlatformItem,
		&itemDetails.PickupPoint.ID,
		&itemDetails.PickupPoint.Address,
		&itemDetails.PickupPoint.City,
		&itemDetails.PickupPoint.IsActive,
		&itemDetails.Owner.ID,
		&itemDetails.Owner.Email,
	)

	if err != nil {
		return nil, err
	}

	itemDetails.City = itemDetails.PickupPoint.City

	return &itemDetails, nil

}

func (im *ItemModel) UpdateByID(itemID string, item ItemUpdate) error {
	query := `
		UPDATE items
		SET
			pickup_point_id = COALESCE($1, pickup_point_id),
			title           = COALESCE($2, title),
			description     = COALESCE($3, description),
			category        = COALESCE($4, category),
			price_per_hour  = COALESCE($5, price_per_hour),
			price_per_day   = COALESCE($6, price_per_day)  
			WHERE id = $7
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{item.PickupPointID, item.Title, item.Description, item.Category, item.PricePerHour, item.PricePerDay, itemID}

	result, err := im.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil

}

func (im *ItemModel) GetByIDRaw(itemID string) (*Item, error) {
	query := `SELECT id, OWNER_ID, PICKUP_POINT_ID, TITLE, DESCRIPTION, CATEGORY, PRICE_PER_HOUR, PRICE_PER_DAY, STATUS,is_platform_item FROM items WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var item Item

	err := im.DB.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID,
		&item.OwnerID,
		&item.PickupPointID,
		&item.Title,
		&item.Description,
		&item.Category,
		&item.PricePerHour,
		&item.PricePerDay,
		&item.Status,
		&item.IsPlatformItem,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (im *ItemModel) DeleteByID(itemID string) error {
	query := `DELETE FROM items WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := im.DB.ExecContext(ctx, query, itemID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil

}

func (im *ItemModel) GetByOwnerID(ownerID string) ([]*ItemList, error) {
	query := `
    SELECT items.id, items.title, items.category, items.price_per_hour, items.price_per_day, pickup_points.city 
    FROM items 
    JOIN pickup_points ON items.pickup_point_id = pickup_points.id 
    WHERE items.owner_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := im.DB.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*ItemList{}
	for rows.Next() {
		var item ItemList
		err := rows.Scan(&item.ID, &item.Title, &item.Category, &item.PricePerHour, &item.PricePerDay, &item.City)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}
