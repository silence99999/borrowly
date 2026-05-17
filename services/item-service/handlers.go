package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

func (app *application) getItems(w http.ResponseWriter, r *http.Request) {
	items, err := app.dbGetItems()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, items)
}

func (app *application) getItemByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := app.dbGetItemByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}
	_ = writeJSON(w, http.StatusOK, item)
}

func (app *application) getMyItems(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	items, err := app.dbGetItemsByOwner(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, items)
}

func (app *application) createItem(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	var in struct {
		PickupPointID string `json:"pickup_point_id"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		Category      string `json:"category"`
		PricePerHour  int    `json:"price_per_hour"`
		PricePerDay   int    `json:"price_per_day"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if _, err := uuid.Parse(in.PickupPointID); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("pickup_point_id must be a valid UUID"))
		return
	}

	item := &Item{
		OwnerID:       u.ID,
		PickupPointID: in.PickupPointID,
		Title:         in.Title,
		Description:   in.Description,
		Category:      in.Category,
		PricePerHour:  in.PricePerHour,
		PricePerDay:   in.PricePerDay,
		IsPlatformItem: u.Role == "ADMIN",
	}
	if err := app.dbCreateItem(item); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	_ = writeJSON(w, http.StatusCreated, item)
}

func (app *application) updateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u := r.Context().Value(authUserKey).(AuthUser)

	item, err := app.dbGetItemRaw(id)
	if err != nil {
		writeError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}
	if item.OwnerID != u.ID && u.Role != "ADMIN" {
		writeError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}

	var in struct {
		PickupPointID *string `json:"pickup_point_id"`
		Title         *string `json:"title"`
		Description   *string `json:"description"`
		Category      *string `json:"category"`
		PricePerHour  *int    `json:"price_per_hour"`
		PricePerDay   *int    `json:"price_per_day"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if in.PickupPointID == nil && in.Title == nil && in.Description == nil &&
		in.Category == nil && in.PricePerHour == nil && in.PricePerDay == nil {
		writeError(w, http.StatusBadRequest, errors.New("no fields to update"))
		return
	}
	if err := app.dbUpdateItem(id, nil, in.PickupPointID, in.Title, in.Description, in.Category, in.PricePerHour, in.PricePerDay); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u := r.Context().Value(authUserKey).(AuthUser)

	item, err := app.dbGetItemRaw(id)
	if err != nil {
		writeError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}
	if item.OwnerID != u.ID && u.Role != "ADMIN" {
		writeError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}
	if err := app.dbDeleteItem(id); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) createItemImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u := r.Context().Value(authUserKey).(AuthUser)

	item, err := app.dbGetItemRaw(id)
	if err != nil {
		writeError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}
	if item.OwnerID != u.ID {
		writeError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	uploadDir := "uploads/item_images"
	_ = os.MkdirAll(uploadDir, os.ModePerm)

	safeName := filepath.Base(header.Filename)
	dst, err := os.Create(filepath.Join(uploadDir, safeName))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	img, err := app.dbCreateItemImage(id, safeName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, img)
}

func (app *application) getPickupPoints(w http.ResponseWriter, r *http.Request) {
	points, err := app.dbGetPickupPoints()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, points)
}

func (app *application) createPickupPoint(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Address  string `json:"address"`
		City     string `json:"city"`
		IsActive bool   `json:"is_active"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	pp := &PickupPoint{Address: in.Address, City: in.City, IsActive: in.IsActive}
	if err := app.dbCreatePickupPoint(pp); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusCreated, pp)
}
