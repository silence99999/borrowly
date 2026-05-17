package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/silence99999/advanced_final/internal/data"
	"github.com/silence99999/advanced_final/internal/json"
)

func (app *application) CreateReviewHandler(w http.ResponseWriter, r *http.Request) {
	itemID, err := app.readIDParam(r)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var input struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}

	err = json.ReadJSON(r, &input)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	reviewedUser := r.Context().Value(authUserKey).(AuthUser)

	if input.Rating < 0 || input.Rating > 5 {
		app.WriteError(w, http.StatusBadRequest, errors.New("rating must be between 0 and 5"))
		return
	}

	review := &data.Review{
		ItemID:         itemID,
		ReviewedUserID: reviewedUser.ID,
		Rating:         input.Rating,
		Comment:        input.Comment,
	}

	err = app.models.Reviews.Create(review)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = json.WriteJSON(w, http.StatusCreated, review)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) GetReviewsHandler(w http.ResponseWriter, r *http.Request) {
	itemID, err := app.readIDParam(r)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	reviews, err := app.models.Reviews.GetAll(itemID)
	if err != nil {
		app.WriteError(w, http.StatusNotFound, err)
		return
	}

	err = json.WriteJSON(w, http.StatusOK, reviews)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) DeleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	itemID, err := app.readIDParam(r)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	reviewID := chi.URLParam(r, "reviewID")
	if reviewID == "" {
		app.WriteError(w, http.StatusBadRequest, errors.New("invalid id parameter"))
		return
	}

	reviewer := r.Context().Value(authUserKey).(AuthUser)

	err = app.models.Reviews.Delete(reviewID, itemID, reviewer.ID)
	if err != nil {
		app.WriteError(w, http.StatusNotFound, err)
		return
	}

	err = json.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Successfully deleted",
	})
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) UpdateReviewHandler(w http.ResponseWriter, r *http.Request) {
	itemID, err := app.readIDParam(r)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	reviewID := chi.URLParam(r, "reviewID")
	if reviewID == "" {
		app.WriteError(w, http.StatusBadRequest, errors.New("invalid id parameter"))
		return
	}

	var input struct {
		Rating  *int    `json:"rating"`
		Comment *string `json:"comment"`
	}

	err = json.ReadJSON(r, &input)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	reviewer := r.Context().Value(authUserKey).(AuthUser)

	if input.Rating != nil && (*input.Rating < 0 || *input.Rating > 5) {
		app.WriteError(w, http.StatusBadRequest, errors.New("rating must be between 0 and 5"))
		return
	}

	if input.Rating == nil && input.Comment == nil {
		app.WriteError(w, http.StatusBadRequest, errors.New("no fields to update"))
		return
	}

	reviewUpdate := data.ReviewUpdate{
		Comment: input.Comment,
		Rating:  input.Rating,
	}

	err = app.models.Reviews.UpdateByID(reviewID, itemID, reviewer.ID, reviewUpdate)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = json.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Successfully updated",
	})
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
	}
}
