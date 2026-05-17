package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func readJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	_ = writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (app *application) getReviews(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	reviews, err := app.dbGetReviews(itemID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, reviews)
}

func (app *application) createReview(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	u := r.Context().Value(authUserKey).(AuthUser)

	var in struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if in.Rating < 0 || in.Rating > 5 {
		writeError(w, http.StatusBadRequest, errors.New("rating must be between 0 and 5"))
		return
	}

	rv := &Review{ItemID: itemID, ReviewedUserID: u.ID, Rating: in.Rating, Comment: in.Comment}
	if err := app.dbCreateReview(rv); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusCreated, rv)
}

func (app *application) updateReview(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	reviewID := chi.URLParam(r, "reviewID")
	u := r.Context().Value(authUserKey).(AuthUser)

	var in struct {
		Rating  *int    `json:"rating"`
		Comment *string `json:"comment"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if in.Rating == nil && in.Comment == nil {
		writeError(w, http.StatusBadRequest, errors.New("no fields to update"))
		return
	}
	if in.Rating != nil && (*in.Rating < 0 || *in.Rating > 5) {
		writeError(w, http.StatusBadRequest, errors.New("rating must be between 0 and 5"))
		return
	}

	if err := app.dbUpdateReview(reviewID, itemID, u.ID, in.Rating, in.Comment); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]string{"message": "updated successfully"})
}

func (app *application) deleteReview(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	reviewID := chi.URLParam(r, "reviewID")
	u := r.Context().Value(authUserKey).(AuthUser)

	if err := app.dbDeleteReview(reviewID, itemID, u.ID); err != nil {
		writeError(w, http.StatusNotFound, errors.New("review not found"))
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]string{"message": "deleted successfully"})
}
