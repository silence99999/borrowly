package data

import (
	"context"
	"database/sql"
	"time"
)

type Review struct {
	ID             string    `json:"id"`
	ItemID         string    `json:"item_id"`
	ReviewedUserID string    `json:"reviewed_user_id"`
	Rating         int       `json:"rating"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"created_at"`
}

type ReviewUpdate struct {
	Rating  *int    `json:"rating"`
	Comment *string `json:"comment"`
}

type ReviewModel struct {
	DB *sql.DB
}

func (rv *ReviewModel) Create(review *Review) error {
	query := `
		INSERT INTO reviews (item_id, reviewed_user_id, rating, comment) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{review.ItemID, review.ReviewedUserID, review.Rating, review.Comment}

	err := rv.DB.QueryRowContext(ctx, query, args...).Scan(&review.ID, &review.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (rv *ReviewModel) GetAll(itemID string) ([]*Review, error) {
	query := `
		SELECT id, item_id, reviewed_user_id, rating, comment, created_at 
		FROM reviews 
		WHERE item_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := rv.DB.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*Review

	for rows.Next() {
		var review Review

		err := rows.Scan(
			&review.ID,
			&review.ItemID,
			&review.ReviewedUserID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (rv *ReviewModel) Delete(reviewID, itemID, reviewerID string) error {
	query := `DELETE FROM reviews WHERE id = $1 AND item_id = $2 AND reviewed_user_id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := rv.DB.ExecContext(ctx, query, reviewID, itemID, reviewerID)
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

func (rv *ReviewModel) UpdateByID(reviewID, itemID, reviewerID string, review ReviewUpdate) error {
	query := `
		UPDATE reviews
		SET
			rating = COALESCE($4, rating),
			comment = COALESCE($5, comment)
		WHERE id = $1 AND item_id = $2 AND reviewed_user_id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{reviewID, itemID, reviewerID, review.Rating, review.Comment}

	result, err := rv.DB.ExecContext(ctx, query, args...)
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
