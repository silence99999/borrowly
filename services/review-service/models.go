package main

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

func (app *application) dbCreateReview(r *Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return app.db.QueryRowContext(ctx,
		`INSERT INTO reviews(item_id,reviewed_user_id,rating,comment) VALUES($1,$2,$3,$4) RETURNING id,created_at`,
		r.ItemID, r.ReviewedUserID, r.Rating, r.Comment,
	).Scan(&r.ID, &r.CreatedAt)
}

func (app *application) dbGetReviews(itemID string) ([]*Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		`SELECT id,item_id,reviewed_user_id,rating,comment,created_at FROM reviews WHERE item_id=$1`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reviews := []*Review{}
	for rows.Next() {
		var rv Review
		if err := rows.Scan(&rv.ID, &rv.ItemID, &rv.ReviewedUserID, &rv.Rating, &rv.Comment, &rv.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, &rv)
	}
	return reviews, rows.Err()
}

func (app *application) dbDeleteReview(reviewID, itemID, reviewerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := app.db.ExecContext(ctx,
		`DELETE FROM reviews WHERE id=$1 AND item_id=$2 AND reviewed_user_id=$3`, reviewID, itemID, reviewerID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (app *application) dbUpdateReview(reviewID, itemID, reviewerID string, rating *int, comment *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := app.db.ExecContext(ctx,
		`UPDATE reviews SET rating=COALESCE($4,rating),comment=COALESCE($5,comment)
		 WHERE id=$1 AND item_id=$2 AND reviewed_user_id=$3`,
		reviewID, itemID, reviewerID, rating, comment)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
