package data

import (
	"context"
	"database/sql"
	"time"
)

type ItemImage struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"item_id"`
	FileName  string    `json:"file_name"`
	CreatedAt time.Time `json:"created_at"`
}

type ItemImageModel struct {
	DB *sql.DB
}

func (iti *ItemImageModel) Create(image *ItemImage) error {
	query := `INSERT INTO item_images (item_id, image_url) VALUES ($1,$2) RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{image.ItemID, image.FileName}

	err := iti.DB.QueryRowContext(ctx, query, args...).Scan(&image.ID)
	if err != nil {
		return err
	}

	return nil

}
