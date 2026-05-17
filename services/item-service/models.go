package main

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
	IsPlatformItem bool      `json:"is_platform_item"`
	CreatedAt      time.Time `json:"created_at"`
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

type ItemDetails struct {
	ID           string      `json:"id"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	Category     string      `json:"category"`
	PricePerHour int         `json:"price_per_hour"`
	PricePerDay  int         `json:"price_per_day"`
	City         string      `json:"pickup_city"`
	PickupPoint  PickupPoint `json:"pickup_point"`
	Owner        OwnerInfo   `json:"owner"`
	IsPlatformItem bool      `json:"is_platform_item"`
}

type OwnerInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type PickupPoint struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	City     string `json:"city"`
	IsActive bool   `json:"is_active"`
}

type ItemImage struct {
	ID       string `json:"id"`
	ItemID   string `json:"item_id"`
	FileName string `json:"file_name"`
}

func (app *application) dbCreateItem(item *Item) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return app.db.QueryRowContext(ctx,
		`INSERT INTO items (owner_id,pickup_point_id,title,description,category,price_per_hour,price_per_day,is_platform_item)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,status`,
		item.OwnerID, item.PickupPointID, item.Title, item.Description,
		item.Category, item.PricePerHour, item.PricePerDay, item.IsPlatformItem,
	).Scan(&item.ID, &item.Status)
}

func (app *application) dbGetItems() ([]*ItemList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		`SELECT i.id,i.title,i.category,i.price_per_hour,i.price_per_day,i.is_platform_item,pp.city
		 FROM items i JOIN pickup_points pp ON i.pickup_point_id=pp.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*ItemList{}
	for rows.Next() {
		var it ItemList
		if err := rows.Scan(&it.ID, &it.Title, &it.Category, &it.PricePerHour, &it.PricePerDay, &it.IsPlatformItem, &it.City); err != nil {
			return nil, err
		}
		items = append(items, &it)
	}
	return items, rows.Err()
}

func (app *application) dbGetItemByID(id string) (*ItemDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var d ItemDetails
	err := app.db.QueryRowContext(ctx,
		`SELECT i.id,i.title,i.description,i.category,i.price_per_hour,i.price_per_day,i.is_platform_item,
		        i.owner_id,pp.id,pp.address,pp.city,pp.is_active
		 FROM items i
		 JOIN pickup_points pp ON i.pickup_point_id=pp.id
		 WHERE i.id=$1`, id,
	).Scan(
		&d.ID, &d.Title, &d.Description, &d.Category, &d.PricePerHour, &d.PricePerDay, &d.IsPlatformItem,
		&d.Owner.ID, &d.PickupPoint.ID, &d.PickupPoint.Address, &d.PickupPoint.City, &d.PickupPoint.IsActive,
	)
	if err != nil {
		return nil, err
	}
	d.City = d.PickupPoint.City
	return &d, nil
}

func (app *application) dbGetItemRaw(id string) (*Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var it Item
	err := app.db.QueryRowContext(ctx,
		`SELECT id,owner_id,pickup_point_id,title,description,category,price_per_hour,price_per_day,status,is_platform_item
		 FROM items WHERE id=$1`, id,
	).Scan(&it.ID, &it.OwnerID, &it.PickupPointID, &it.Title, &it.Description,
		&it.Category, &it.PricePerHour, &it.PricePerDay, &it.Status, &it.IsPlatformItem)
	return &it, err
}

func (app *application) dbUpdateItem(id string, ownerID *string, pickupPointID, title, description, category *string, pricePerHour, pricePerDay *int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := app.db.ExecContext(ctx,
		`UPDATE items SET
			pickup_point_id=COALESCE($1,pickup_point_id),
			title=COALESCE($2,title),
			description=COALESCE($3,description),
			category=COALESCE($4,category),
			price_per_hour=COALESCE($5,price_per_hour),
			price_per_day=COALESCE($6,price_per_day)
		 WHERE id=$7`,
		pickupPointID, title, description, category, pricePerHour, pricePerDay, id,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (app *application) dbDeleteItem(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := app.db.ExecContext(ctx, `DELETE FROM items WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (app *application) dbGetItemsByOwner(ownerID string) ([]*ItemList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		`SELECT i.id,i.title,i.category,i.price_per_hour,i.price_per_day,i.is_platform_item,pp.city
		 FROM items i JOIN pickup_points pp ON i.pickup_point_id=pp.id WHERE i.owner_id=$1`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*ItemList{}
	for rows.Next() {
		var it ItemList
		if err := rows.Scan(&it.ID, &it.Title, &it.Category, &it.PricePerHour, &it.PricePerDay, &it.IsPlatformItem, &it.City); err != nil {
			return nil, err
		}
		items = append(items, &it)
	}
	return items, rows.Err()
}

func (app *application) dbCreateItemImage(itemID, fileName string) (*ItemImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	img := &ItemImage{ItemID: itemID, FileName: fileName}
	err := app.db.QueryRowContext(ctx,
		`INSERT INTO item_images(item_id,image_url) VALUES($1,$2) RETURNING id`, itemID, fileName,
	).Scan(&img.ID)
	return img, err
}

func (app *application) dbGetPickupPoints() ([]*PickupPoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		`SELECT id,address,city,is_active FROM pickup_points WHERE is_active=true ORDER BY city,address`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	points := []*PickupPoint{}
	for rows.Next() {
		var pp PickupPoint
		if err := rows.Scan(&pp.ID, &pp.Address, &pp.City, &pp.IsActive); err != nil {
			return nil, err
		}
		points = append(points, &pp)
	}
	return points, rows.Err()
}

func (app *application) dbCreatePickupPoint(pp *PickupPoint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return app.db.QueryRowContext(ctx,
		`INSERT INTO pickup_points(address,city,is_active) VALUES($1,$2,$3) RETURNING id`,
		pp.Address, pp.City, pp.IsActive,
	).Scan(&pp.ID)
}
